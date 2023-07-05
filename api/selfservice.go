package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/application-research/delta-dm/core"
	db "github.com/application-research/delta-dm/db"
	"github.com/labstack/echo/v4"
)

const PROVIDER = "PROVIDER"

type SelfServiceResponse struct {
	Cid             string `json:"cid"`
	ContentLocation string `json:"content_location"`
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

	selfService.PUT("/telemetry/:cid", func(c echo.Context) error {
		return handleSelfServiceTelemetry(c, dldm)
	})

	selfService.GET("/eligible_pieces", func(c echo.Context) error {
		return handleSelfServiceEligiblePieces(c, dldm)
	})
}

func selfServiceTokenMiddleware(dldm *core.DeltaDM) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providerToken := c.Request().Header.Get("X-DELTA-AUTH")

			if providerToken == "" {
				return c.String(401, "missing provider self-service token")
			}
			var p db.Provider
			res := dldm.DB.Model(&db.Provider{}).Preload("ReplicationProfiles").Where("key = ?", providerToken).Find(&p)

			if res.Error != nil {
				log.Errorf("error finding provider: %s", res.Error)
				return c.String(401, "unable to find provider for self-service token")
			}
			if p.ActorID == "" {
				return c.String(401, "invalid provider self-service token")
			}
			if !p.AllowSelfService {
				return c.String(401, "provider is not allowed to self-serve, please contact administrator to enable it")
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
	endEpochAdvance := c.QueryParam("end_epoch_advance")

	var delayDays uint64 = DEFAULT_DELAY_DAYS
	var advanceDays uint64 = 0

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

	if endEpochAdvance != "" {
		var err error
		advanceDays, err = strconv.ParseUint(endEpochAdvance, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse end_epoch_advance: %s", err)
		}

		if advanceDays < 0 || advanceDays > 20 {
			return fmt.Errorf("end_epoch_advance must be between 0 and 20 days")
		}
	}

	if piece == "" {
		return fmt.Errorf("must provide a piece CID")
	}

	p := c.Get(PROVIDER).(db.Provider)

	var cnt db.Content
	res := dldm.DB.Model(&db.Content{}).Preload("Replications").Where("comm_p = ?", piece).Find(&cnt)
	if res.Error != nil {
		return fmt.Errorf("unable to make deal for this CID")
	}

	var ds db.Dataset
	res = dldm.DB.Model(&db.Dataset{}).Where("id = ?", cnt.DatasetID).Find(&ds)
	if res.Error != nil {
		return fmt.Errorf("unable to find dataset %d associated with requested CID", cnt.DatasetID)
	}

	var rp db.ReplicationProfile
	isAllowed := false
	for _, thisRp := range p.ReplicationProfiles {
		if thisRp.DatasetID == ds.ID {
			isAllowed = true
			rp = thisRp
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
		if repl.ProviderActorID == p.ActorID && !repl.Status.HasFailed() {
			return fmt.Errorf("content '%s' is already replicated to provider '%s'", piece, p.ActorID)
		}
	}

	var dealsToMake core.OfflineDealRequest
	log.Debugf("calling DELTA api for deal\n\n")

	wallet, err := walletSelection(dldm.DB, &cnt.DatasetID)

	if err != nil || wallet.Addr == "" {
		return fmt.Errorf("dataset '%s' does not have a wallet. no deals were made. please contact administrator", ds.Name)
	}

	dealsToMake = append(dealsToMake, core.Deal{
		PayloadCID: cnt.PayloadCID,
		Wallet: db.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode:     "import",
		Miner:              p.ActorID,
		Size:               cnt.Size,
		SkipIpniAnnounce:   !rp.Indexed,
		RemoveUnsealedCopy: !rp.Unsealed,
		DurationInDays:     ds.DealDuration - advanceDays,
		StartEpochInDays:   delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        cnt.CommP,
			PaddedPieceSize: cnt.PaddedSize,
		},
	})

	_, err = dldm.MakeDeals(dealsToMake, dldm.DAPI.ServiceAuthToken, true)
	if err != nil {
		return fmt.Errorf("unable to make deal for this CID: %s", err)
	}

	return c.JSON(http.StatusOK, SelfServiceResponse{Cid: cnt.CommP, ContentLocation: cnt.ContentLocation})
}

func handleSelfServiceByDataset(c echo.Context, dldm *core.DeltaDM) error {
	dataset := c.Param("dataset")
	startEpochDelay := c.QueryParam("start_epoch_delay")
	endEpochAdvance := c.QueryParam("end_epoch_advance")

	if dataset == "" {
		return fmt.Errorf("must provide a dataset name")
	}

	var ds db.Dataset
	dsRes := dldm.DB.Where("name = ?", dataset).First(&ds)
	if dsRes.Error != nil || ds.ID == 0 {
		return fmt.Errorf("invalid dataset: %s", dsRes.Error)
	}

	var advanceDays uint64 = 0
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

	if endEpochAdvance != "" {
		var err error
		advanceDays, err = strconv.ParseUint(endEpochAdvance, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse end_epoch_advance: %s", err)
		}

		if advanceDays < 0 || advanceDays > 20 {
			return fmt.Errorf("end_epoch_advance must be between 0 and 20 days")
		}
	}

	p := c.Get(PROVIDER).(db.Provider)

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
	cnt, err := findUnreplicatedContentForProvider(dldm.DB, p.ActorID, &ds.ID, &numDeals, false)
	if err != nil {
		return fmt.Errorf("unable to find content for dataset: %s", err)
	}

	if len(cnt) == 0 {
		return fmt.Errorf("no deals available for dataset")
	}

	deal := cnt[0]

	wallet, err := walletSelection(dldm.DB, &ds.ID)

	if err != nil || wallet.Addr == "" {
		return fmt.Errorf("dataset '%s' does not have a wallet associated. no deals were made. please contact administrator", ds.Name)
	}

	var dealsToMake []core.Deal

	dealsToMake = append(dealsToMake, core.Deal{
		PayloadCID: deal.PayloadCID,
		Wallet: db.Wallet{
			Addr: wallet.Addr,
		},
		ConnectionMode:     "import",
		Miner:              p.ActorID,
		Size:               deal.Size,
		SkipIpniAnnounce:   !deal.Indexed,
		RemoveUnsealedCopy: !deal.Unsealed,
		DurationInDays:     deal.DealDuration - advanceDays,
		StartEpochInDays:   delayDays,
		PieceCommitment: core.PieceCommitment{
			PieceCid:        deal.CommP,
			PaddedPieceSize: deal.PaddedSize,
		},
	})

	_, err = dldm.MakeDeals(dealsToMake, dldm.DAPI.ServiceAuthToken, true)
	if err != nil {
		return fmt.Errorf("unable to make deal for this CID: %s", err)
	}

	return c.JSON(http.StatusOK, SelfServiceResponse{Cid: deal.CommP, ContentLocation: deal.ContentLocation})
}

type SelfServiceStatusUpdate struct {
	DealUuid string `json:"deal_uuid"`
	State    string `json:"state"`
	Message  string `json:"message"`
}

func handleSelfServiceTelemetry(c echo.Context, dldm *core.DeltaDM) error {
	var update SelfServiceStatusUpdate
	if err := c.Bind(&update); err != nil {
		return fmt.Errorf("unable to bind request: %s", err)
	}

	if update.DealUuid == "" {
		return fmt.Errorf("must provide a deal_uuid")
	}

	var repl db.Replication
	err := dldm.DB.Model(&repl).Where("deal_uuid = ?", update.DealUuid).First(&repl).Error

	if err != nil {
		return fmt.Errorf("unable to find content for CID: %s", err)
	}

	p := c.Get(PROVIDER).(db.Provider)
	if repl.ProviderActorID != p.ActorID {
		return fmt.Errorf("deal '%s' does not belong to provider '%s'", update.DealUuid, p.ActorID)
	}

	repl.SelfService.LastUpdate = time.Now()
	repl.SelfService.Status = update.State
	repl.SelfService.Message = update.Message

	err = dldm.DB.Save(&repl).Error
	if err != nil {
		return fmt.Errorf("unable to update deal status: %s", err)
	}

	return nil
}

type EligiblePiece struct {
	PayloadCID      string `json:"payload_cid"`
	PieceCID        string `json:"piece_cid"`
	Size            uint64 `json:"size"`
	PaddedSize      uint64 `json:"padded_size"`
	ContentLocation string `json:"content_location"`
}

func handleSelfServiceEligiblePieces(c echo.Context, dldm *core.DeltaDM) error {
	p := c.Get(PROVIDER).(db.Provider)
	limit := c.QueryParam("limit")

	var numDeals uint
	if limit != "" {
		n, err := strconv.ParseUint(limit, 10, 64)

		if err != nil {
			return fmt.Errorf("unable to parse limit: %s", err)
		}

		if numDeals > 2000 {
			return fmt.Errorf("limit must be less than 2000")
		}

		numDeals = uint(n)
	} else {
		numDeals = uint(500)
	}

	cnt, err := findUnreplicatedContentForProvider(dldm.DB, p.ActorID, nil, &numDeals, true)

	if err != nil {
		return fmt.Errorf("unable to find content for dataset: %s", err)
	}

	var result []EligiblePiece

	for _, deal := range cnt {
		result = append(result, EligiblePiece{
			PayloadCID:      deal.PayloadCID,
			PieceCID:        deal.CommP,
			Size:            deal.Size,
			PaddedSize:      deal.PaddedSize,
			ContentLocation: deal.ContentLocation,
		})
	}

	return c.JSON(http.StatusOK, result)
}
