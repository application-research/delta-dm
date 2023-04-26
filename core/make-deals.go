package core

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/application-research/delta-dm/util"
	"gorm.io/gorm"
)

// Make deals for the given OfflineDealRequests, and update DDM database accordingly
func (dldm *DeltaDM) MakeDeals(dealsToMake OfflineDealRequest, authKey string, isSelfService bool) (*OfflineDealResponse, error) {
	if dldm.DryRunMode {
		fmt.Println(util.Red + "-- DRY RUN MODE (NO DEALS MADE) --" + util.Reset)
		fmt.Printf("\n\n %+v \n\n", dealsToMake)
		fmt.Println(util.Red + "---------------------------------" + util.Reset)
		return &OfflineDealResponse{}, nil
	}

	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake, authKey)
	if err != nil {
		return nil, fmt.Errorf("unable to make deal with delta api: %s", err)
	}

	for _, c := range *deltaResp {
		if c.Status != "success" {
			continue
		}
		var newReplication = Replication{
			ContentCommP:    c.DealRequestMeta.PieceCommitment.PieceCid,
			ProviderActorID: c.DealRequestMeta.Miner,
			DeltaContentID:  c.DeltaContentID,
			DealTime:        time.Now(),
			Status:          StatusPending,
			IsSelfService:   isSelfService,
			ProposalCid:     "PENDING_" + fmt.Sprint(rand.Int()),
		}

		res := dldm.DB.Model(&Replication{}).Create(&newReplication)
		if res.Error != nil {
			log.Errorf("unable to create replication in db: %s", res.Error)
			continue
		}

		// Update the content's num replications
		dldm.DB.Model(&Content{}).Where("comm_p = ?", newReplication.ContentCommP).Update("num_replications", gorm.Expr("num_replications + ?", 1))
	}
	return deltaResp, nil
}
