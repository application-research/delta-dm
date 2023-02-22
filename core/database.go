package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	//	"gorm.io/gorm/logger"

	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("router")
)

func OpenDatabase(dbName string) (*gorm.DB, error) {
	DB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		//	Logger: logger.Default.LogMode(logger.Info),
	})

	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	err := db.AutoMigrate(&Replication{}, &Provider{}, &Dataset{}, &Content{})
	if err != nil {
		log.Fatalf("error migrating database: %s", err)
	}
}

// A replication refers to a deal, for a specific content, with a client
type Replication struct {
	gorm.Model
	Content         Content   `json:"content"`
	DealTime        time.Time `json:"deal_time"`
	ProposalCid     string    `json:"proposal_cid" gorm:"unique"`
	ProviderActorID string    `json:"provider_actor_id"`
	ContentCommP    string    `json:"content_commp"`
}

// A client is a Storage Provider that is being replicated to
type Provider struct {
	Key          uuid.UUID     `json:"key" gorm:"type:uuid"`
	ActorID      string        `json:"actor_id" gorm:"primaryKey"`
	Replications []Replication `json:"replications" gorm:"foreignKey:ProviderActorID"`
}

// A Dataset is a collection of CAR files, and is identified by a name/slug
type Dataset struct {
	gorm.Model
	Name             string    `json:"name" gorm:"unique; not null"`
	ReplicationQuota int       `json:"replication_quota"`
	DealDuration     int       `json:"deal_duration"`
	Wallet           string    `json:"wallet"`
	Unsealed         bool      `json:"unsealed"`
	Indexed          bool      `json:"indexed"`
	Contents         []Content `json:"contents" gorm:"foreignKey:DatasetID"`
}

type Content struct {
	CommP           string `json:"commp" gorm:"primaryKey"`
	Size            int64  `json:"size"`
	PaddedSize      int64  `json:"padded_size"`
	DatasetID       int
	Replications    []Replication `json:"replications" gorm:"foreignKey:ContentCommP"`
	NumReplications int           `json:"num_replications"`
}
