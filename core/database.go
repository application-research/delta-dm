package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//	"gorm.io/gorm/logger"

	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("router")
)

// Opens a database connection, and returns a gorm DB object.
// It will automatically detect either Postgres DSN or, or will fallback to sqlite
func OpenDatabase(dbDsn string, debug bool) (*gorm.DB, error) {
	var DB *gorm.DB
	var err error
	var config = &gorm.Config{}
	if debug {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	if dbDsn[:8] == "postgres" {
		DB, err = gorm.Open(postgres.Open(dbDsn), config)
	} else {
		DB, err = gorm.Open(sqlite.Open(dbDsn), config)
	}

	// generate new models.
	ConfigureModels(DB) // create models.

	if debug {
		log.Debugf("connected to db at: %s", dbDsn)
	}

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	err := db.AutoMigrate(&Provider{}, &Dataset{}, &Content{}, &Wallet{}, &ReplicationProfile{}, &WalletDatasets{}, &Replication{})

	if err != nil {
		log.Fatalf("error migrating database: %s", err)
	}
}

type DealStatus string

const (
	DealStatusPending DealStatus = "PENDING"
	DealStatusSuccess DealStatus = "SUCCESS"
	DealStatusFailure DealStatus = "FAILURE"
)

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
	Content         Content    `json:"content"` //TODO: doesnt come back from api
	DealTime        time.Time  `json:"deal_time"`
	DeltaContentID  int64      `json:"delta_content_id" gorm:"unique"`
	DealUUID        string     `json:"deal_uuid" gorm:"unique"`
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
