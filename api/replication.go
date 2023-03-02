package api

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/application-research/delta-dm/core"
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

const EPOCHS_PER_DAY = 2880 // 86400s / 30s = 2880 epochs per day

func ConfigureReplicationRouter(e *echo.Group, dldm *core.DeltaDM) {
	replication := e.Group("/replication")

	replication.GET("", func(c echo.Context) error {

		p := c.QueryParam("provider")
		ds := c.QueryParam("dataset")

		var r []core.Replication

		tx := dldm.DB.Model(&core.Replication{}).Joins("Content")

		if ds != "" {
			tx.Where("Content.dataset_name = ?", ds)
		}

		if p != "" {
			tx.Where("replications.provider_actor_id = ?", p)
		}

		tx.Find(&r)

		return c.JSON(200, r)
	})

	replication.POST("", func(c echo.Context) error {
		return handlePostReplication(c, dldm)
	})

}

// POST /api/replication
// @param num number of deals requested
// @returns a slice of the CIDs
func handlePostReplication(c echo.Context, dldm *core.DeltaDM) error {
	var d PostReplicationBody

	err := RequestAuthHeaderCheck(c)
	if err != nil {
		return c.JSON(401, err.Error())
	}
	authorizationString := c.Request().Header.Get("Authorization")

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
		if c.Wallet.Addr == "" {
			return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please add a wallet for this dataset and try again. alternatively, explicitly specify a dataset in the request to force replication of one with an existing wallet", c.Dataset.Name)
		}

		dealsToMake = append(dealsToMake, core.Deal{
			Cid: c.PayloadCID, // Payload CID
			Wallet: core.Wallet{
				Addr: c.Wallet.Addr,
			},
			ConnectionMode:       "import",
			Miner:                d.Provider,
			Size:                 c.Size,
			SkipIpniAnnounce:     !c.Indexed,
			RemoveUnsealedCopies: !c.Unsealed,
			// TODO: DealDuration needs to be capped based on start epoch, hardcoded to 7 for now
			Duration: int64((c.DealDuration - 7) * EPOCHS_PER_DAY),
			// TODO: start epoch
			PieceCommitment: core.PieceCommitment{
				PieceCid:        c.CommP,
				PaddedPieceSize: c.PaddedSize,
			},
		})
	}

	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake, authorizationString)
	if err != nil {
		return fmt.Errorf("unable to make deal with delta api: %s", err)
	}

	for _, c := range *deltaResp {
		if c.Status != "success" {
			continue
		}
		var newReplication = core.Replication{
			ContentCommP:    c.RequestMeta.PieceCommitment.PieceCid,
			ProviderActorID: c.RequestMeta.Miner,
			DeltaContentID:  c.ContentID,
			DealTime:        time.Now(),
			ProposalCid:     "PENDING_" + fmt.Sprint(rand.Int()), // TODO: From delta
		}

		res := dldm.DB.Model(&core.Replication{}).Create(&newReplication)
		if res.Error != nil {
			log.Errorf("unable to create replication in db: %s", res.Error)
			continue
		}

		// Update the content's num replications
		dldm.DB.Model(&core.Content{}).Where("comm_p = ?", newReplication.ContentCommP).Update("num_replications", gorm.Expr("num_replications + ?", 1))

	}

	return c.JSON(200, deltaResp)
}

type replicatedContentQueryResponse struct {
	core.Content
	core.Dataset
	core.Wallet
}

// Query the database for all contant that does not have replications to this actor yet
// Arguments: providerID - the actor ID of the provider
// 					  datasetName (optional) - the name of the dataset to replicate
// 					  numDeals (optional) - the number of replications (deals) to return. If nil, return all
func findUnreplicatedContentForProvider(db *gorm.DB, providerID string, datasetName *string, numDeals *uint) ([]replicatedContentQueryResponse, error) {

	rawQuery := "select * from datasets d inner join contents c " +
		"on d.name = c.dataset_name inner join wallets w on d.name = w.dataset_name where c.comm_p not in " +
		"(select r.content_comm_p from replications r where r.provider_actor_id not in (select p.actor_id from providers p where p.actor_id not in (?))) " +
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
