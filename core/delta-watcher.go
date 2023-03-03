package core

import (
	"fmt"
	"time"

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
	go watch(ddm.DB, ddm.DAPI)
}

func watch(db *gorm.DB, d *DeltaAPI) {
	for {
		err := RunReconciliation(db, d)

		if err != nil {
			log.Errorf("failed running delta reconciliation job: %s", err)
		}
		time.Sleep(30 * time.Second)
	}
}

func RunReconciliation(db *gorm.DB, d *DeltaAPI) error {
	log.Debug("starting reconcile task")
	var pendingReplications []int64

	db.Model(&Replication{}).Where("status = ?", StatusPending).Select("delta_content_id").Find(&pendingReplications)

	log.Debugf("reconciling %d pending replications", len(pendingReplications))
	statsResponse, err := d.GetDealStatus(pendingReplications)
	if err != nil {
		return fmt.Errorf("could not get deal status: %s", err)
	}

	ru := computeReplicationUpdates(*statsResponse)

	log.Debugf("updating %d replications", len(ru))
	for _, r := range ru {
		err := db.Model(&Replication{}).Where("delta_content_id = ?", r.DeltaContentID).Updates(r)

		if err.Error != nil {
			return fmt.Errorf("could not update replication: %s", err.Error)
		}
	}

	return nil
}

func computeReplicationUpdates(dealStats DealStatsResponse) []Replication {
	toUpdate := []Replication{}

	for _, deal := range dealStats {
		switch deal.Content.Status {

		// Success!
		case CONTENT_DEAL_PROPOSAL_SENT:
			toUpdate = append(toUpdate, Replication{
				Status:         StatusSuccess,
				ProposalCid:    deal.DealProposals[0].Signed, // TODO: @alvin-reyes is this the right place to get proposal cid?
				DeltaContentID: deal.Content.ID,
				DeltaMessage:   deal.Content.LastMessage,
			})
		case CONTENT_DEAL_PROPOSAL_FAILED:
			toUpdate = append(toUpdate, Replication{
				Status:         StatusFailure,
				ProposalCid:    deal.DealProposals[0].Signed, // TODO: @alvin-reyes is this the right place to get proposal cid?
				DeltaContentID: deal.Content.ID,
				DeltaMessage:   deal.Content.LastMessage,
			})

			// ? Do we need to care about any other statuses?
		}
	}

	return toUpdate
}
