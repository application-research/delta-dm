package db

import (
	"time"

	sm "github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DealStatus string

const DDM_StorageDealStatusPending DealStatus = "DDM_PENDING"

// List of statuses that indicate a deal has failed
var FailedStatuses = []DealStatus{
	DealStatus(sm.DealStates[sm.StorageDealProposalRejected]),
	DealStatus(sm.DealStates[sm.StorageDealError]),
	// These events come from Delta, if the deal never makes it to the chain
	DealStatus("transfer-failed"),
	DealStatus("deal-proposal-failed"),
}

func (ds DealStatus) HasFailed() bool {
	for _, state := range FailedStatuses {
		if state == ds {
			return true
		}
	}
	return false
}

// This is separate from the `DealStatus` enum to accomodate more granular statuses in the future (ex, SealingInProgress)
type SelfServiceStatus string

const (
	SelfServiceStatusPending SelfServiceStatus = "PENDING"
	SelfServiceStatusSuccess SelfServiceStatus = "SUCCESS"
	SelfServiceStatusFailure SelfServiceStatus = "FAILURE"
)

// A replication refers to a deal, for a specific content, with a client
type Replication struct {
	gorm.Model
	Content         Content    `json:"content"`
	DealTime        time.Time  `json:"deal_time"`
	DeltaContentID  int64      `json:"delta_content_id" gorm:"unique"`
	DealUUID        string     `json:"deal_uuid"`
	OnChainDealID   uint       `json:"on_chain_deal_id"`
	ProposalCid     string     `json:"proposal_cid" gorm:"unique"`
	ProviderActorID string     `json:"provider_actor_id"`
	ContentCommP    string     `json:"content_commp"`
	Status          DealStatus `json:"status" gorm:"notnull,default:'PENDING'"`
	DeltaMessage    string     `json:"delta_message,omitempty"`
	SelfService     struct {
		IsSelfService bool      `json:"is_self_service"`
		LastUpdate    time.Time `json:"last_update"`
		Status        string    `json:"status" gorm:"notnull,default:'PENDING'"`
		Message       string    `json:"message"`
	} `json:"self_service" gorm:"embedded;embeddedPrefix:ss_"`
}

// A client is a Storage Provider that is being replicated to
type Provider struct {
	Key                 uuid.UUID            `json:"key,omitempty" gorm:"type:uuid"`
	ActorID             string               `json:"actor_id" gorm:"primaryKey"`
	ActorName           string               `json:"actor_name,omitempty"`
	AllowSelfService    bool                 `json:"allow_self_service,omitempty" gorm:"notnull,default:true"`
	BytesReplicated     ByteSizes            `json:"bytes_replicated,omitempty" gorm:"-"`
	CountReplicated     uint64               `json:"count_replicated,omitempty" gorm:"-"`
	Replications        []Replication        `json:"replications,omitempty" gorm:"foreignKey:ProviderActorID"`
	ReplicationProfiles []ReplicationProfile `json:"replication_profiles" gorm:"foreignKey:ProviderActorID"`
}

type ReplicationProfile struct {
	ProviderActorID string `gorm:"primaryKey;uniqueIndex:idx_provider_dataset" json:"provider_actor_id"`
	DatasetID       uint   `gorm:"primaryKey;uniqueIndex:idx_provider_dataset" json:"dataset_id"`
	Unsealed        bool   `json:"unsealed"`
	Indexed         bool   `json:"indexed"`
}

type ByteSizes struct {
	Raw    uint64 `json:"raw"`
	Padded uint64 `json:"padded"`
}

// A Dataset is a collection of CAR files, and is identified by a name/slug
type Dataset struct {
	gorm.Model
	Name                string               `json:"name" gorm:"unique; not null"`
	ReplicationQuota    uint64               `json:"replication_quota"`
	DealDuration        uint64               `json:"deal_duration"`
	Wallets             []Wallet             `json:"wallets,omitempty" gorm:"many2many:wallet_datasets;"`
	Contents            []Content            `json:"contents" gorm:"foreignKey:DatasetID;references:ID"`
	BytesReplicated     ByteSizes            `json:"bytes_replicated,omitempty" gorm:"-"`
	BytesTotal          ByteSizes            `json:"bytes_total,omitempty" gorm:"-"`
	CountReplicated     uint64               `json:"count_replicated" gorm:"-"`
	CountTotal          uint64               `json:"count_total" gorm:"-"`
	ReplicationProfiles []ReplicationProfile `json:"replication_profiles" gorm:"foreignKey:dataset_id"`
}

type Content struct {
	CommP           string        `json:"commp" csv:"commP" gorm:"primaryKey"`
	PayloadCID      string        `json:"payload_cid" csv:"payloadCid"`
	Size            uint64        `json:"size" csv:"size"`
	PaddedSize      uint64        `json:"padded_size" csv:"paddedSize"`
	DatasetID       uint          `json:"dataset_id"`
	Replications    []Replication `json:"replications,omitempty" gorm:"foreignKey:ContentCommP"`
	NumReplications uint64        `json:"num_replications"`
	ContentLocation string        `json:"content_location"`
}

type WalletDatasets struct {
	WalletAddr string `gorm:"primaryKey" json:"wallet_addr"`
	DatasetID  uint   `gorm:"primaryKey" json:"dataset_id"`
}

type Wallet struct {
	Addr     string        `json:"address" gorm:"primaryKey"`
	Datasets []Dataset     `json:"datasets" gorm:"many2many:wallet_datasets;"`
	Balance  WalletBalance `json:"balance,omitempty" gorm:"-"`
}

type WalletBalance struct {
	BalanceFilecoin uint64 `json:"balance_filecoin"`
	BalanceDatacap  uint64 `json:"balance_datacap"`
}
