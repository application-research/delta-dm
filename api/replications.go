package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/application-research/delta-dm/core"
	"github.com/application-research/delta-dm/util"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type PostReplicationBody struct {
	Provider string  `json:"provider"`
	Dataset  *string `json:"dataset,omitempty"`
	NumDeals *uint   `json:"num_deals,omitempty"`
	// NumTib       *int    `json:"num_tib,omitempty"`
	PricePerDeal float64 `json:"price_per_deal,omitempty"`
}

func ConfigureReplicationsRouter(e *echo.Group, dldm *core.DeltaDM) {
	replications := e.Group("/replications")

	replications.Use(dldm.AS.AuthMiddleware)

	replications.GET("", func(c echo.Context) error {
		return handleGetReplications(c, dldm)
	})

	replications.POST("", func(c echo.Context) error {
		return handlePostReplications(c, dldm)
	})

}

type GetReplicationsQueryParams struct {
	Statuses      []string
	Datasets      []string
	Providers     []string
	SelfService   *bool
	DealTimeStart *time.Time
	DealTimeEnd   *time.Time
	ProposalCid   *string
	PieceCid      *string
	Message       *string
	Limit         int
	Offset        int
}

// Extract all the replications query parameters from the request
func extractGetReplicationsQueryParams(c echo.Context) GetReplicationsQueryParams {
	var gqp GetReplicationsQueryParams

	proposalCid := c.QueryParam("proposal_cid")
	pieceCid := c.QueryParam("piece_cid")
	statuses := c.QueryParam("statuses")
	datasets := c.QueryParam("datasets")
	providers := c.QueryParam("providers")
	selfService := c.QueryParam("self_service")
	dealTimeStart := c.QueryParam("deal_time_start")
	dealTimeEnd := c.QueryParam("deal_time_end")
	message := c.QueryParam("message")
	limit := c.QueryParam("limit")
	offset := c.QueryParam("offset")

	var err error
	gqp.Limit, err = strconv.Atoi(limit)
	if err != nil {
		gqp.Limit = 100
	}

	gqp.Offset, err = strconv.Atoi(offset)
	if err != nil {
		gqp.Offset = 0
	}

	// PieceCID and ProposalCID will result in a specific search, so can return them right away
	if pieceCid != "" {
		gqp.PieceCid = &pieceCid
		return gqp
	}
	if proposalCid != "" {
		gqp.ProposalCid = &proposalCid
		return gqp
	}

	if statuses != "" {
		gqp.Statuses = strings.Split(strings.ToUpper(statuses), ",")
	}

	if datasets != "" {
		gqp.Datasets = strings.Split(datasets, ",")
	}

	if providers != "" {
		gqp.Providers = strings.Split(providers, ",")
	}

	if message != "" {
		gqp.Message = &message
	}

	ss, err := strconv.ParseBool(selfService)
	if err == nil && selfService != "" {
		gqp.SelfService = &ss
	}

	dts, err := util.EpochStringToTime(dealTimeStart)
	if err == nil {
		gqp.DealTimeStart = &dts
	}

	dte, err := util.EpochStringToTime(dealTimeEnd)
	if err == nil {
		gqp.DealTimeEnd = &dte
	}

	return gqp
}

// handleGetReplications handles the request to get replications
// @Summary Get replications
// @Tags replications
// @Produce  json
func handleGetReplications(c echo.Context, dldm *core.DeltaDM) error {
	rqp := extractGetReplicationsQueryParams(c)

	var r []core.Replication

	tx := dldm.DB.Model(&core.Replication{}).Joins("Content")

	if rqp.PieceCid != nil {
		tx.Where("replications.piece_cid = ?", rqp.PieceCid)
	} else if rqp.ProposalCid != nil {
		tx.Where("replications.proposal_cid = ?", rqp.ProposalCid)
	}

	if len(rqp.Statuses) > 0 {
		tx.Where("replications.status IN ?", rqp.Statuses)
	}

	if len(rqp.Datasets) > 0 {
		tx.Where("Content.dataset_name IN ?", rqp.Datasets)
	}

	if len(rqp.Providers) > 0 {
		tx.Where("replications.provider_actor_id IN ?", rqp.Providers)
	}

	if rqp.SelfService != nil {
		tx.Where("replications.is_self_service = ?", rqp.SelfService)
	}

	if rqp.DealTimeStart != nil {
		tx.Where("replications.deal_time >= ?", rqp.DealTimeStart)
	}

	if rqp.DealTimeEnd != nil {
		tx.Where("replications.deal_time <= ?", rqp.DealTimeEnd)
	}

	if rqp.Message != nil {
		tx.Where("replications.delta_message LIKE ?", "%"+*rqp.Message+"%")
	}

	tx.Limit(rqp.Limit).Offset(rqp.Offset).Order("replications.id DESC").Scan(&r)

	return c.JSON(200, r)
}

// POST /api/replication
// @param num number of deals requested
// @returns a slice of the CIDs
func handlePostReplications(c echo.Context, dldm *core.DeltaDM) error {
	var d PostReplicationBody

	authKey := c.Get(core.AUTH_KEY).(string)

	if err := c.Bind(&d); err != nil {
		return err
	}

	if d.NumDeals == nil {
		return fmt.Errorf("must specify num_deals")
	}

	// TODO: Support num_tib to allow specifying the amount of data to replicate

	toReplicate, err := findUnreplicatedContentForProvider(dldm.DB, d.Provider, d.Dataset, d.NumDeals)
	if err != nil {
		return err
	}

	var dealsToMake core.OfflineDealRequest
	log.Debugf("calling DELTA api for %+v deals\n\n", len(toReplicate))

	for _, c := range toReplicate {
		wallet, err := walletSelection(dldm.DB, &c.DatasetName)

		if err != nil || wallet.Addr == "" {
			return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please add a wallet for this dataset and try again. alternatively, explicitly specify a dataset in the request to force replication of one with an existing wallet", c.Dataset.Name)
		}

		dealsToMake = append(dealsToMake, core.Deal{
			Cid: c.PayloadCID,
			Wallet: core.Wallet{
				Addr: wallet.Addr,
			},
			ConnectionMode:       "import",
			Miner:                d.Provider,
			Size:                 c.Size,
			SkipIpniAnnounce:     !c.Indexed,
			RemoveUnsealedCopies: !c.Unsealed,
			DurationInDays:       c.DealDuration - 3, // TODO: Allow specifying duration, with default
			StartEpochAtDays:     3,
			PieceCommitment: core.PieceCommitment{
				PieceCid:        c.CommP,
				PaddedPieceSize: c.PaddedSize,
			},
		})
	}

	deltaResp, err := dldm.MakeDeals(dealsToMake, authKey, false)
	if err != nil {
		return fmt.Errorf("unable to make deals: %s", err)
	}

	return c.JSON(200, deltaResp)
}

type replicatedContentQueryResponse struct {
	core.Content
	core.Dataset
}

// Query the database for all contant that does not have replications to this actor yet
// Arguments: providerID - the actor ID of the provider
// 					  datasetName (optional) - the name of the dataset to replicate
// 					  numDeals (optional) - the number of replications (deals) to return. If nil, return all
func findUnreplicatedContentForProvider(db *gorm.DB, providerID string, datasetName *string, numDeals *uint) ([]replicatedContentQueryResponse, error) {

	rawQuery := "select * from datasets d inner join contents c " +
		"on d.name = c.dataset_name where c.comm_p not in " +
		"(select r.content_comm_p from replications r where r.status != 'FAILURE' and r.provider_actor_id not in (select p.actor_id from providers p where p.actor_id not in (?))) " +
		"and c.num_replications < d.replication_quota"
	var rawValues = []interface{}{providerID}

	if datasetName != nil {
		rawQuery += " AND d.name = ?"
		rawValues = append(rawValues, datasetName)
	}

	if numDeals != nil {
		rawQuery += " LIMIT ?"
		rawValues = append(rawValues, numDeals)
	}
	var contents []replicatedContentQueryResponse
	db.Raw(rawQuery, rawValues...).Scan(&contents)

	return contents, nil
}

// Find which wallet to use when making deals for a given dataset
func walletSelection(db *gorm.DB, datasetName *string) (*core.Wallet, error) {
	var w []core.Wallet
	res := db.Model(&core.Wallet{}).Where("dataset_name = ?", datasetName).Find(&w)

	if res.Error != nil {
		return nil, res.Error
	}

	if len(w) == 0 {
		return nil, fmt.Errorf("no wallet found for dataset '%s'", *datasetName)

	}

	// TODO: Wallet selection algorithm
	// Just choose the first wallet for now
	return &w[0], nil
}