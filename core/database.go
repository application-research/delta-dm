package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//	"gorm.io/gorm/logger"

	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("router")
)

func OpenDatabase(dbName string, debug bool) (*gorm.DB, error) {
	var config = &gorm.Config{}
	if debug {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	DB, err := gorm.Open(sqlite.Open(dbName), config)

	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	err := db.AutoMigrate(&Replication{}, &Provider{}, &Dataset{}, &Content{}, &Wallet{})
	if err != nil {
		log.Fatalf("error migrating database: %s", err)
	}
}

type ReplicationStatus string

const (
	StatusPending ReplicationStatus = "PENDING"
	StatusSuccess ReplicationStatus = "SUCCESS"
	StatusFailure ReplicationStatus = "FAILURE"
)

// A replication refers to a deal, for a specific content, with a client
type Replication struct {
	gorm.Model
	Content         Content           `json:"content"` //TODO: doesnt come back from api
	DealTime        time.Time         `json:"deal_time"`
	DeltaContentID  int64             `json:"delta_content_id" gorm:"unique"`
	ProposalCid     string            `json:"proposal_cid" gorm:"unique"`
	ProviderActorID string            `json:"provider_actor_id"`
	ContentCommP    string            `json:"content_commp"`
	Status          ReplicationStatus `json:"status" gorm:"notnull,default:'PENDING'"`
	DeltaMessage    string            `json:"delta_message,omitempty"`
}

// A client is a Storage Provider that is being replicated to
type Provider struct {
	Key             uuid.UUID     `json:"key" gorm:"type:uuid"`
	ActorID         string        `json:"actor_id" gorm:"primaryKey"`
	ActorName       string        `json:"actor_name"`
	BytesReplicated ByteSizes     `json:"bytes_replicated,omitempty" gorm:"-"`
	Replications    []Replication `json:"replications" gorm:"foreignKey:ProviderActorID"`
}

type ByteSizes struct {
	Raw    uint64 `json:"raw"`
	Padded uint64 `json:"padded"`
}

// A Dataset is a collection of CAR files, and is identified by a name/slug
type Dataset struct {
	gorm.Model
	Name             string    `json:"name" gorm:"unique; not null"`
	ReplicationQuota uint      `json:"replication_quota"`
	DealDuration     uint64    `json:"deal_duration"`
	Wallet           []Wallet  `json:"wallet,omitempty" gorm:"foreignKey:DatasetName;references:Name"`
	Unsealed         bool      `json:"unsealed"`
	Indexed          bool      `json:"indexed"`
	Contents         []Content `json:"contents" gorm:"foreignKey:DatasetName;references:Name"`
	BytesReplicated  ByteSizes `json:"bytes_replicated,omitempty" gorm:"-"`
	BytesTotal       ByteSizes `json:"bytes_total,omitempty" gorm:"-"`
}

type Content struct {
	CommP           string        `json:"commp" csv:"commP" gorm:"primaryKey"`
	PayloadCID      string        `json:"payload_cid" csv:"payloadCid"`
	Size            uint64        `json:"size" csv:"size"`
	PaddedSize      uint64        `json:"padded_size" csv:"paddedSize"`
	DatasetName     string        `json:"dataset_name"`
	Replications    []Replication `json:"replications,omitempty" gorm:"foreignKey:ContentCommP"`
	NumReplications uint          `json:"num_replications"`
}

type Wallet struct {
	Addr        string        `json:"address" gorm:"primaryKey"`
	DatasetName string        `json:"dataset_name"`
	Balance     WalletBalance `json:"balance,omitempty" gorm:"-"`
}

type WalletBalance struct {
	BalanceFilecoin uint64 `json:"balance_filecoin"`
	BalanceDatacap  uint64 `json:"balance_datacap"`
}
