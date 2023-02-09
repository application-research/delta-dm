package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDatabase() (*gorm.DB, error) {
	dbName, okHost := viper.Get("DB_NAME").(string)
	if !okHost {
		panic("DB_NAME not set")
	}
	DB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Replication{}, &Provider{}, &Dataset{}, &Carfile{})
}

type DealState string

// const (
// 	PENDING  DealState = "PENDING"  // Deal has been made
// 	COMPLETE DealState = "COMPLETE" // Deal is successfully onchain
// 	FAILED   DealState = "FAILED"   // Deal failed
// )

// A replication refers to a deal, for a specific carfile, with a client
type Replication struct {
	gorm.Model
	client  Provider
	carfile Carfile
	// state    DealState // TODO: directly from delta core?
	dealTime    time.Time
	proposalCid string // TODO: type
}

// A client is a Storage Provider that is being replicated to
type Provider struct {
	gorm.Model
	key     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	actorID string    `gorm:"unique,not null"`
}

// A Dataset is a collection of CAR files, and is identified by a slug
type Dataset struct {
	gorm.Model
	name             string
	dealLength       int64 // num. epochs
	replicationQuota int
	carfiles         []Carfile
}

type Carfile struct {
	commp       string `gorm:"primarykey"`
	size        int64
	padded_size int64
}
