package core

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDatabase(dbName string) (*gorm.DB, error) {
	DB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})

	// generate new models.
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

// A replication refers to a deal, for a specific carfile, with a client
type Replication struct {
	gorm.Model
	client  Provider
	content Content
	// state    DealState // TODO: directly from delta core?
	dealTime    time.Time
	proposalCid string // TODO: type
}

// A client is a Storage Provider that is being replicated to
type Provider struct {
	gorm.Model
	Key     uuid.UUID `json:"key" gorm:"type:uuid"`
	ActorID string    `json:"actor_id" gorm:"unique; not null"`
}

// A Dataset is a collection of CAR files, and is identified by a slug
type Dataset struct {
	gorm.Model
	Name             string `json:"name" gorm:"unique; not null"`
	ReplicationQuota int    `json:"replication_quota"`
	DealDuration     int    `json:"deal_duration"`
	Wallet           string `json:"wallet"`
	Unsealed         bool   `json:"unsealed"`
	Indexed          bool   `json:"indexed"`
	contents         []Content
}

type Content struct {
	gorm.Model
	commp       string `gorm:"primaryKey"`
	size        int64
	padded_size int64
}
