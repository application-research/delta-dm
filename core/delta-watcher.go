package core

import (
	"fmt"
	"time"

	db "github.com/application-research/delta-dm/db"
	"github.com/application-research/delta-dm/util"
	"gorm.io/gorm"
)

// TODO: Import from Delta once public
// https://github.com/application-research/delta/blob/main/utils/constants.go
const (
	CONTENT_PIECE_COMPUTING        = "piece-computing"
	CONTENT_PIECE_COMPUTED         = "piece-computed"
	CONTENT_PIECE_COMPUTING_FAILED = "piece-computing-failed"
	CONTENT_PIECE_ASSIGNED         = "piece-assigned"

	CONTENT_DEAL_MAKING_PROPOSAL  = "making-deal-proposal"
	CONTENT_DEAL_SENDING_PROPOSAL = "sending-deal-proposal"
	CONTENT_DEAL_PROPOSAL_SENT    = "deal-proposal-sent"
	CONTENT_DEAL_PROPOSAL_FAILED  = "deal-proposal-failed"

	DEAL_STATUS_TRANSFER_STARTED  = "transfer-started"
	DEAL_STATUS_TRANSFER_FINISHED = "transfer-finished"
	DEAL_STATUS_TRANSFER_FAILED   = "transfer-failed"
)

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

	dbi.Model(&db.Replication{}).Where("status = ?", db.DealStatusPending).Select("delta_content_id").Find(&pendingReplications)

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
		if r.Status == db.DealStatusFailure {
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
		switch deal.Content.Status {

		// Success!
		case CONTENT_DEAL_PROPOSAL_SENT:
			r := db.Replication{
				Status:         db.DealStatusSuccess,
				DeltaContentID: deal.Content.ID,
				DeltaMessage:   deal.Content.LastMessage,
			}
			if len(deal.Deals) > 0 {
				r.ProposalCid = deal.Deals[0].PropCid
				r.DealUUID = deal.Deals[0].DealUUID
				r.OnChainDealID = deal.Deals[0].DealID
			}
			if len(deal.PieceCommitments) > 0 {
				r.ContentCommP = deal.PieceCommitments[0].Piece
			}

			toUpdate = append(toUpdate, r)

		case CONTENT_DEAL_PROPOSAL_FAILED:
			r := db.Replication{
				Status:         db.DealStatusFailure,
				DeltaContentID: deal.Content.ID,
				DeltaMessage:   deal.Content.LastMessage,
			}
			if len(deal.PieceCommitments) > 0 {
				r.ContentCommP = deal.PieceCommitments[0].Piece
			}

			if len(deal.Deals) > 0 {
				r.ProposalCid = deal.Deals[0].PropCid
				r.DealUUID = deal.Deals[0].DealUUID
				r.OnChainDealID = deal.Deals[0].DealID
			}
			toUpdate = append(toUpdate, r)

			// ? Do we need to care about any other statuses?
		}
	}

	return toUpdate
}
