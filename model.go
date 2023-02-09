package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DealState string

const (
	PENDING  DealState = "PENDING"  // Deal has been made
	COMPLETE DealState = "COMPLETE" // Deal is successfully onchain
	FAILED   DealState = "FAILED"   // Deal failed
)

// A replication refers to a deal, for a specific carfile, with a client
type Replication struct {
	gorm.Model
	client   Client
	carfile  Carfile
	state    DealState
	dealTime time.Time
}

// A client is a Storage Provider that is being replicated to
type Client struct {
	gorm.Model
	key     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	actorID string    `gorm:"unique,not null"`
}

// A Dataset is a collection of CAR files, and is identified by a slug
type Dataset struct {
	gorm.Model
	slug     string
	carfiles []Carfile
}

type Carfile struct {
	commp string `gorm:"primarykey"`
}
