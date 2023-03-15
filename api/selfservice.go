package api

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type SelfServiceRequestBody struct {
	// ProviderToken string  `json:"key"`
	// Dataset  *string `json:"dataset,omitempty"`
	PieceCid string `json:"piece_cid,omitempty"`
}

func ConfigureSelfServiceRouter(e *echo.Group, dldm *core.DeltaDM) {
	selfService := e.Group("/self-service")

	selfService.GET("/by-cid/:piece", func(c echo.Context) error {
		return handleSelfServicePostByCid(c, dldm)
	})

}

// POST /api/self-service/
// @param
// @returns a slice of the CIDs
func handleSelfServicePostByCid(c echo.Context, dldm *core.DeltaDM) error {
	piece := c.Param("piece")
	startEpochDelay := c.QueryParam("start_epoch_delay")
	var delayDays uint64 = 3

	if startEpochDelay != "" {
		var err error
		delayDays, err = strconv.ParseUint(startEpochDelay, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse start_epoch_delay: %s", err)
		}

		if delayDays < 1 || delayDays > 14 {
			return fmt.Errorf("start_epoch_delay must be between 1 and 14 days")
		}
	}

	if piece == "" {
		return fmt.Errorf("must provide a piece CID")
	}

	providerToken := c.Request().Header.Get("Authorization")

	var p core.Provider
	res := dldm.DB.Model(&core.Provider{}).Where("key = ?", providerToken).Find(&p)

	if res.Error != nil {
		log.Errorf("error finding provider: %s", res.Error)
		return fmt.Errorf("unable to find provider for token")
	}

	cnt, err := findContentByCommP(dldm.DB, p.ActorID, piece)
	if err != nil {
		return fmt.Errorf("unable to make deal for this CID")
	}

	var dealsToMake core.OfflineDealRequest
	log.Debugf("calling DELTA api for deal\n\n")

	wallet, err := walletSelection(dldm.DB, &cnt.DatasetName)

	if err != nil || wallet.Addr == "" {
		return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please add a wallet for this dataset and try again. alternatively, explicitly specify a dataset in the request to force replication of one with an existing wallet", cnt.DatasetName)
	}

	dealsToMake = append(dealsToMake, core.Deal{
		Cid: cnt.PayloadCID,
		Wallet: core.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode:       "import",
		Miner:                p.ActorID,
		Size:                 cnt.Size,
		SkipIpniAnnounce:     !cnt.Indexed,
		RemoveUnsealedCopies: !cnt.Unsealed,
		DurationInDays:       cnt.DealDuration,
		StartEpochAtDays:     delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        cnt.CommP,
			PaddedPieceSize: cnt.PaddedSize,
		},
	})

	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake, providerToken)
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
			Status:          core.StatusPending,
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

func findContentByCommP(db *gorm.DB, providerID string, commp string) (replicatedContentQueryResponse, error) {

	rawQuery := "select * from datasets d inner join contents c " +
		"on d.name = c.dataset_name where c.comm_p not in " +
		"(select r.content_comm_p from replications r where r.status != 'FAILURE' and r.provider_actor_id not in (select p.actor_id from providers p where p.actor_id not in (?))) " +
		"AND c.num_replications < d.replication_quota AND c.commp = ? LIMIT 1" //TODO: cleanup query!
	var rawValues = []interface{}{providerID, commp}

	var content replicatedContentQueryResponse
	db.Raw(rawQuery, rawValues...).Scan(&content)

	return content, nil
}
