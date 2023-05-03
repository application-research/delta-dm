package api

import (
	"fmt"
	"strconv"

	"github.com/application-research/delta-dm/core"
	"github.com/labstack/echo/v4"
)

const PROVIDER = "PROVIDER"

type SelfServiceResponse struct {
	Cid string `json:"cid"`
}

func ConfigureSelfServiceRouter(e *echo.Group, dldm *core.DeltaDM) {
	selfService := e.Group("/self-service")

	selfService.Use(selfServiceTokenMiddleware(dldm))

	selfService.GET("/by-cid/:piece", func(c echo.Context) error {
		return handleSelfServiceByCid(c, dldm)
	})

	selfService.GET("/by-dataset/:dataset", func(c echo.Context) error {
		return handleSelfServiceByDataset(c, dldm)
	})
}

func selfServiceTokenMiddleware(dldm *core.DeltaDM) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providerToken := c.Request().Header.Get("X-DELTA-AUTH")

			if providerToken == "" {
				return c.String(401, "missing provider self-service token")
			}
			var p core.Provider
			res := dldm.DB.Model(&core.Provider{}).Preload("ReplicationProfiles").Where("key = ?", providerToken).Find(&p)

			if res.Error != nil {
				log.Errorf("error finding provider: %s", res.Error)
				return c.String(401, "unable to find provider for self-service token")
			}
			if p.ActorID == "" {
				return c.String(401, "invalid provider self-service token")
			}

			c.Set(PROVIDER, p)

			return next(c)
		}
	}
}

// POST /api/self-service/by-cid/:piece
// @param :piece Piece CID of content to replicate
// @queryparam
// @returns a slice of the CIDs
func handleSelfServiceByCid(c echo.Context, dldm *core.DeltaDM) error {
	piece := c.Param("piece")
	startEpochDelay := c.QueryParam("start_epoch_delay")
	var delayDays uint64 = DEFAULT_DELAY_DAYS

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

	p := c.Get(PROVIDER).(core.Provider)

	var cnt core.Content
	res := dldm.DB.Model(&core.Content{}).Preload("Replications").Where("comm_p = ?", piece).Find(&cnt)
	if res.Error != nil {
		return fmt.Errorf("unable to make deal for this CID")
	}

	var ds core.Dataset
	res = dldm.DB.Model(&core.Dataset{}).Where("name = ?", cnt.DatasetName).Find(&ds)
	if res.Error != nil {
		return fmt.Errorf("unable to find associated dataset %s", cnt.DatasetName)
	}

	isAllowed := false
	for _, rp := range p.ReplicationProfiles {
		if rp.DatasetID == ds.ID {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("provider '%s' is not allowed to replicate dataset '%s'", p.ActorID, ds.Name)
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
		PayloadCID: cnt.PayloadCID,
		Wallet: core.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode: "import",
		Miner:          p.ActorID,
		Size:           cnt.Size,
		// SkipIpniAnnounce:   !ds.Indexed,
		// RemoveUnsealedCopy: !ds.Unsealed,
		DurationInDays:   ds.DealDuration,
		StartEpochInDays: delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        cnt.CommP,
			PaddedPieceSize: cnt.PaddedSize,
		},
	})

	_, err = dldm.MakeDeals(dealsToMake, dldm.DAPI.ServiceAuthToken, true)
	if err != nil {
		return fmt.Errorf("unable to make deal for this CID: %s", err)
	}

	return c.JSON(200, SelfServiceResponse{Cid: cnt.CommP})
}

func handleSelfServiceByDataset(c echo.Context, dldm *core.DeltaDM) error {
	dataset := c.Param("dataset")
	startEpochDelay := c.QueryParam("start_epoch_delay")

	if dataset == "" {
		return fmt.Errorf("must provide a dataset name")
	}

	var ds core.Dataset
	dsRes := dldm.DB.Where("name = ?", dataset).First(&ds)
	if dsRes.Error != nil || ds.ID == 0 {
		return fmt.Errorf("invalid dataset: %s", dsRes.Error)
	}

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

	p := c.Get(PROVIDER).(core.Provider)

	fmt.Printf("\n\n%+v\n\n", p)

	isAllowed := false
	for _, rp := range p.ReplicationProfiles {
		if rp.DatasetID == ds.ID {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("provider '%s' is not allowed to replicate dataset '%s'", p.ActorID, dataset)
	}

	// give one deal at a time
	numDeals := uint(1)
	cnt, err := findUnreplicatedContentForProvider(dldm.DB, p.ActorID, &dataset, &numDeals)
	if err != nil {
		return fmt.Errorf("unable to find content for dataset: %s", err)
	}

	if len(cnt) == 0 {
		return fmt.Errorf("no deals available for dataset")
	}

	deal := cnt[0]

	wallet, err := walletSelection(dldm.DB, &deal.DatasetName)

	if err != nil || wallet.Addr == "" {
		return fmt.Errorf("dataset '%s' does not have a wallet associated. no deals were made. please contact administrator", deal.DatasetName)
	}

	var dealsToMake []core.Deal

	dealsToMake = append(dealsToMake, core.Deal{
		PayloadCID: deal.PayloadCID,
		Wallet: core.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode: "import",
		Miner:          p.ActorID,
		Size:           deal.Size,
		// SkipIpniAnnounce:   !deal.Indexed,
		// RemoveUnsealedCopy: !deal.Unsealed,
		DurationInDays:   deal.DealDuration - delayDays,
		StartEpochInDays: delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        deal.CommP,
			PaddedPieceSize: deal.PaddedSize,
		},
	})

	_, err = dldm.MakeDeals(dealsToMake, dldm.DAPI.ServiceAuthToken, true)
	if err != nil {
		return fmt.Errorf("unable to make deal for this CID: %s", err)
	}

	return c.JSON(200, SelfServiceResponse{Cid: deal.CommP})
}
