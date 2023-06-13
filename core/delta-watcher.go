package core

import (
	"fmt"
	"time"

	db "github.com/application-research/delta-dm/db"
	"github.com/application-research/delta-dm/util"
	"gorm.io/gorm"
)

// TODO: Import from Delta once public

func (ddm *DeltaDM) WatchReplications() {
	if ddm.DryRunMode {
		fmt.Println(util.Red + "disabling Delta watcher in dry run mode" + util.Reset)
		return
	}
	go watch(ddm.DB, ddm.DAPI)
}

func watch(db *gorm.DB, d *DeltaAPI) {
	for {
		time.Sleep(10 * time.Second)

		err := RunReconciliation(db, d)

		if err != nil {
			log.Errorf("failed running delta reconciliation job: %s", err)
		}
	}
}

func RunReconciliation(dbi *gorm.DB, d *DeltaAPI) error {
	log.Debug("starting reconcile task")
	var pendingReplications []int64

	// Once the on_chain_deal_id is nonzero, we don't need to continue checking the deal
	dbi.Model(&db.Replication{}).Where("on_chain_deal_id = ?", 0).Select("delta_content_id").Find(&pendingReplications)

	if len(pendingReplications) == 0 {
		log.Debug("no pending replications")
		return nil
	}

	log.Debugf("reconciling %v\n", pendingReplications)
	statsResponse, err := d.GetDealStatus(pendingReplications)
	if err != nil {
		return fmt.Errorf("could not get deal status: %s", err)
	}

	ru := computeReplicationUpdates(*statsResponse)

	log.Debugf("updating %d replications\n", len(ru))
	for _, r := range ru {
		err := dbi.Model(&db.Replication{}).Where("delta_content_id = ?", r.DeltaContentID).Updates(r)

		if err.Error != nil {
			return fmt.Errorf("could not update replication: %s", err.Error)
		}

		// Remove a replication if it failed
		if r.Status.HasFailed() {
			var cnt db.Content

			err := dbi.Model(&db.Content{}).Where("comm_p = ?", r.ContentCommP).First(&cnt)
			if err.Error != nil {
				return fmt.Errorf("could not find associated content: %s", err.Error)
			}
			// This condition should always be true, but just in case
			if cnt.NumReplications > 0 {
				cnt.NumReplications -= 1
			}

			err = dbi.Save(&cnt)
			if err.Error != nil {
				return fmt.Errorf("could not update associated content: %s", err.Error)
			}
		}

	}

	return nil
}

func computeReplicationUpdates(dealStats DealStatsResponse) []db.Replication {
	toUpdate := []db.Replication{}

	for _, deal := range dealStats {

		r := db.Replication{
			Status:         deal.Content.Status,
			DeltaContentID: deal.Content.ID,
			DeltaMessage:   deal.Content.LastMessage,
		}

		if len(deal.Deals) > 0 {
			r.ProposalCid = deal.Deals[0].PropCid
			r.DealUUID = deal.Deals[0].DealUUID

			if deal.Deals[0].DealID != 0 {
				r.OnChainDealID = deal.Deals[0].DealID
			}

		}
		if len(deal.PieceCommitments) > 0 {
			r.ContentCommP = deal.PieceCommitments[0].Piece
		}

		toUpdate = append(toUpdate, r)
	}

	return toUpdate
}
