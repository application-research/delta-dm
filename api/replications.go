package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/application-research/delta-dm/core"
	db "github.com/application-research/delta-dm/db"
	"github.com/application-research/delta-dm/util"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const DEFAULT_DELAY_DAYS = 3

type PostReplicationBody struct {
	Provider       string  `json:"provider"`
	DatasetID      *uint   `json:"dataset_id,omitempty"`
	NumDeals       *uint   `json:"num_deals,omitempty"`
	DelayStartDays *uint64 `json:"delay_start_days,omitempty"`
	// NumTib       *int    `json:"num_tib,omitempty"`
	// PricePerDeal float64 `json:"price_per_deal,omitempty"`
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

type ReplicationResponse struct {
	Data       []db.Replication `json:"data"`
	TotalCount int64            `json:"totalCount"`
}

// handleGetReplications handles the request to get replications
// @Summary Get replications
// @Tags replications
// @Produce  json
func handleGetReplications(c echo.Context, dldm *core.DeltaDM) error {
	rqp := extractGetReplicationsQueryParams(c)

	tx := dldm.DB.Model(&db.Replication{}).Joins("Content")

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
		tx.Where("replications.ss_is_self_service = ?", rqp.SelfService)
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

	var r []db.Replication
	var totalCount int64

	// Clone the tx before counting
	// GORM's .Count() method does not include JOIN operations because it's designed to optimize counting rows directly from the target table for performance reasons.
	// When you call .Count(), it modifies the current query to remove all selection fields, JOIN clauses, ORDER BY, LIMIT and OFFSET, and replaces it with SELECT count(*) FROM your_table.
	countTx := tx.Session(&gorm.Session{NewDB: false})
	countTx.Count(&totalCount)

	tx.Limit(rqp.Limit).Offset(rqp.Offset).Order("replications.id DESC").Scan(&r)

	response := ReplicationResponse{
		Data:       r,
		TotalCount: totalCount,
	}

	return c.JSON(http.StatusOK, response)
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

	var providerExists bool
	err := dldm.DB.Model(db.Provider{}).
		Select("count(*) > 0").
		Where("actor_id = ?", d.Provider).
		Find(&providerExists).
		Error

	if err != nil {
		return fmt.Errorf("could not check if provider %s exists: %s", d.Provider, err)
	}

	if !providerExists {
		return fmt.Errorf("provider %s does not exist in ddm. please add it first", d.Provider)
	}

	var delayStartEpoch uint64 = DEFAULT_DELAY_DAYS
	if d.DelayStartDays != nil {
		if *d.DelayStartDays < 1 || *d.DelayStartDays > 14 {
			return fmt.Errorf("delay_start_epoch must be between 1 and 14")
		}
		delayStartEpoch = *d.DelayStartDays
	}

	if d.DatasetID != nil {
		var datasetExists bool
		err = dldm.DB.Model(db.Dataset{}).
			Select("count(*) > 0").
			Where("id = ?", d.DatasetID).
			Find(&datasetExists).
			Error
		if err != nil {
			return fmt.Errorf("could not check if dataset with id %d exists: %s", *d.DatasetID, err)
		}
		if !datasetExists {
			return fmt.Errorf("dataset id %d does not exist in ddm.", *d.DatasetID)
		}
	}

	// TODO: Support num_tib to allow specifying the amount of data to replicate

	toReplicate, err := findUnreplicatedContentForProvider(dldm.DB, d.Provider, d.DatasetID, d.NumDeals, false)
	if err != nil {
		return err
	}

	if len(toReplicate) == 0 {
		return fmt.Errorf("no content to replicate to this provider was found. check dataset-provider allowances, replication quota")
	}

	var dealsToMake core.OfflineDealRequest
	log.Debugf("calling DELTA api for %+v deals\n\n", len(toReplicate))

	for _, c := range toReplicate {
		wallet, err := walletSelection(dldm.DB, &c.DatasetID)

		if err != nil || wallet.Addr == "" {
			return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please add a wallet for this dataset and try again. alternatively, explicitly specify a dataset in the request to force replication of one with an existing wallet", c.Dataset.Name)
		}

		dealsToMake = append(dealsToMake, core.Deal{
			PayloadCID: c.PayloadCID,
			Wallet: db.Wallet{
				Addr: wallet.Addr,
			},
			ConnectionMode:     "import",
			Miner:              d.Provider,
			Size:               c.Size,
			SkipIpniAnnounce:   !c.Indexed,
			RemoveUnsealedCopy: !c.Unsealed,
			DurationInDays:     c.DealDuration,
			StartEpochInDays:   delayStartEpoch,
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

	return c.JSON(http.StatusOK, deltaResp)
}

type replicatedContentQueryResponse struct {
	db.Content
	db.Dataset
	// Note: We can't use `db.ReplicationProfile` here because it has a `DatasetID` field which conflicts with the `Dataset` field above
	// Thus, Unsealed and Indexed are added manually
	Unsealed bool
	Indexed  bool
}

// Query the database for all contant that does not have replications to this actor yet
// Arguments: providerID - the actor ID of the provider
//
//	datasetID (optional) - the ID of the dataset to replicate
//	numDeals (optional) - the number of replications (deals) to return. If nil, return all
//  filterOnlyContentLocations - if true, only return content where the content_location is present (i.e, downloadable)
func findUnreplicatedContentForProvider(db *gorm.DB, providerID string, datasetId *uint, numDeals *uint, filterOnlyContentLocations bool) ([]replicatedContentQueryResponse, error) {

	rawQuery := `
  SELECT *
  FROM datasets d
  INNER JOIN contents c ON d.id = c.dataset_id
  INNER JOIN replication_profiles rp ON rp.dataset_id = d.id
	-- Only select content that does not have a non-failed replication to this provider
  WHERE c.comm_p NOT IN (
    SELECT r.content_comm_p 
    FROM replications r 
    WHERE r.status != 'FAILURE' 
    AND r.provider_actor_id NOT IN (
      SELECT p.actor_id 
      FROM providers p 
      WHERE p.actor_id <> ?
    )
  )
  -- Only select content from datasets that this provider is allowed to replicate
  AND rp.provider_actor_id = ?
  AND c.num_replications < d.replication_quota 
	`

	if filterOnlyContentLocations {
		rawQuery += " AND c.content_location NOT NULL"
	}
	var rawValues = []interface{}{providerID, providerID}

	if datasetId != nil && *datasetId != 0 {
		rawQuery += " AND d.id = ?"
		rawValues = append(rawValues, datasetId)
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
func walletSelection(dbi *gorm.DB, datasetId *uint) (*db.Wallet, error) {
	var w []db.Wallet

	res := dbi.Raw("select * from wallets w inner join wallet_datasets wd on w.addr = wd.wallet_addr inner join datasets d on wd.dataset_id = d.id where d.id = ?", datasetId).Scan(&w)

	if res.Error != nil {
		return nil, res.Error
	}

	if len(w) == 0 {
		return nil, fmt.Errorf("no wallet found for dataset '%d'", *datasetId)

	}

	// TODO: Wallet selection algorithm
	// Just choose the first wallet for now
	return &w[0], nil
}
