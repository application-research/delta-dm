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

func ConfigureSelfServiceRouter(e *echo.Group, dldm *core.DeltaDM) {
	selfService := e.Group("/self-service")

	selfService.GET("/by-cid/:piece", func(c echo.Context) error {
		return handleSelfServicePostByCid(c, dldm)
	})

}

// POST /api/self-service/by-cid/:piece
// @param :piece Piece CID of content to replicate
// @queryparam
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

	providerToken := c.Request().Header.Get("X-DELTA-AUTH")

	var p core.Provider
	res := dldm.DB.Model(&core.Provider{}).Where("key = ?", providerToken).Find(&p)

	if res.Error != nil {
		log.Errorf("error finding provider: %s", res.Error)
		return fmt.Errorf("unable to find provider for token")
	}

	if p.ActorID == "" {
		return fmt.Errorf("invalid delta auth token")
	}

	fmt.Printf("%+v\n\n", p)

	// cnt, err := findContentByCommP(dldm.DB, p.ActorID, piece)
	var cnt core.Content
	res = dldm.DB.Model(&core.Content{}).Preload("Replications").Where("comm_p = ?", piece).Find(&cnt)
	if res.Error != nil {
		return fmt.Errorf("unable to make deal for this CID")
	}

	var ds core.Dataset
	res = dldm.DB.Model(&core.Dataset{}).Where("name = ?", cnt.DatasetName).Find(&ds)
	if res.Error != nil {
		return fmt.Errorf("unable to find associated dataset %s", cnt.DatasetName)
	}

	if cnt.NumReplications >= ds.ReplicationQuota {
		return fmt.Errorf("content '%s' has reached its replication quota of %d", piece, ds.ReplicationQuota)
	}

	// Ensure no pending/successful replications have been made for this content to this provider
	for _, repl := range cnt.Replications {
		if repl.ProviderActorID == p.ActorID && repl.Status != core.StatusFailure {
			return fmt.Errorf("content '%s' is already replicated to provider '%s'", piece, p.ActorID)
		}
	}

	var dealsToMake core.OfflineDealRequest
	log.Debugf("calling DELTA api for deal\n\n")

	wallet, err := walletSelection(dldm.DB, &cnt.DatasetName)

	if err != nil || wallet.Addr == "" {
		return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please contact administrator", cnt.DatasetName)
	}

	dealsToMake = append(dealsToMake, core.Deal{
		Cid: cnt.PayloadCID,
		Wallet: core.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode:       "import",
		Miner:                p.ActorID,
		Size:                 cnt.Size,
		SkipIpniAnnounce:     !ds.Indexed,
		RemoveUnsealedCopies: !ds.Unsealed,
		DurationInDays:       ds.DealDuration - delayDays,
		StartEpochAtDays:     delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        cnt.CommP,
			PaddedPieceSize: cnt.PaddedSize,
		},
	})

	// TODO: dont access the token directly here
	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake, dldm.DAPI.ServiceAuthToken)
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
			ProposalCid:     "PENDING_" + fmt.Sprint(rand.Int()),
		}

		res := dldm.DB.Model(&core.Replication{}).Create(&newReplication)
		if res.Error != nil {
			log.Errorf("unable to create replication in db: %s", res.Error)
			continue
		}

		// Update the content's num replications
		dldm.DB.Model(&core.Content{}).Where("comm_p = ?", newReplication.ContentCommP).Update("num_replications", gorm.Expr("num_replications + ?", 1))

	}

	return c.JSON(200, fmt.Sprintf("successfully made deal with %s", p.ActorID))
}
