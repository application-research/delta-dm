package core

import (
	"fmt"
	"math/rand"
	"time"

	db "github.com/application-research/delta-dm/db"
	"github.com/application-research/delta-dm/util"
	"gorm.io/gorm"
)

// Make deals for the given OfflineDealRequests, and update DDM database accordingly
func (dldm *DeltaDM) MakeDeals(dealsToMake OfflineDealRequest, authKey string, isSelfService bool) (*OfflineDealResponse, error) {
	if dldm.DryRunMode {
		fmt.Println(util.Red + "-- DRY RUN MODE (NO DEALS MADE) --" + util.Reset)
		fmt.Printf("\n\n %+v \n\n", dealsToMake)
		fmt.Println(util.Red + "---------------------------------" + util.Reset)

		dealResp, _ := dryRunDeal(&dealsToMake)

		for _, c := range *dealResp {
			var newReplication = db.Replication{
				ContentCommP:    c.DealRequestMeta.PieceCommitment.PieceCid,
				ProviderActorID: c.DealRequestMeta.Miner,
				DeltaContentID:  c.DeltaContentID,
				DealTime:        time.Now(),
				Status:          db.DealStatusSuccess,
				OnChainDealID:   0,
				ProposalCid:     "DRY_RUN_" + fmt.Sprint(rand.Int()),
				DealUUID:        "DRY_RUN_" + fmt.Sprint(rand.Int()),
				DeltaMessage:    "this is a dry run, no deal was made",
			}
			newReplication.SelfService.IsSelfService = isSelfService

			res := dldm.DB.Model(&db.Replication{}).Create(&newReplication)
			if res.Error != nil {
				log.Errorf("unable to create replication in db: %s", res.Error)
				continue
			}
		}

		return dealResp, nil
	}

	deltaResp, err := dldm.DAPI.MakeOfflineDeals(dealsToMake, authKey)
	if err != nil {
		return nil, fmt.Errorf("unable to make deal with delta api: %s", err)
	}

	for _, c := range *deltaResp {
		if c.Status != "success" {
			continue
		}
		var newReplication = db.Replication{
			ContentCommP:    c.DealRequestMeta.PieceCommitment.PieceCid,
			ProviderActorID: c.DealRequestMeta.Miner,
			DeltaContentID:  c.DeltaContentID,
			DealTime:        time.Now(),
			Status:          db.DealStatusPending,
			OnChainDealID:   0,
			ProposalCid:     "PENDING_" + fmt.Sprint(rand.Int()),
			DealUUID:        "PENDING_" + fmt.Sprint(rand.Int()),
		}
		newReplication.SelfService.IsSelfService = isSelfService

		res := dldm.DB.Model(&db.Replication{}).Create(&newReplication)
		if res.Error != nil {
			log.Errorf("unable to create replication in db: %s", res.Error)
			continue
		}

		// Update the content's num replications
		dldm.DB.Model(&db.Content{}).Where("comm_p = ?", newReplication.ContentCommP).Update("num_replications", gorm.Expr("num_replications + ?", 1))
	}
	return deltaResp, nil
}

// Stub function to generate a mocked deal response for local testing.
func dryRunDeal(odr *OfflineDealRequest) (*OfflineDealResponse, error) {
	var resp OfflineDealResponse

	for _, d := range *odr {
		var newResponse = OfflineDealResponseElement{
			DealRequestMeta: d,
			DeltaContentID:  rand.Int63(),
			Status:          "success",
		}

		resp = append(resp, newResponse)
	}

	return &resp, nil
}
