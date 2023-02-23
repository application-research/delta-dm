package core

import (
	"gorm.io/gorm"
)

type DeltaDM struct {
	DAPI *DeltaAPI
	DB   *gorm.DB
}

func NewDeltaDM(dbConnStr string, deltaApi string, deltaAuthToken string) *DeltaDM {
	db, err := OpenDatabase(dbConnStr)
	if err != nil {
		log.Fatalf("could not connect to db: %s", err)
	} else {
		log.Debugf("successfully connected to delta api at %s\n", deltaApi)
	}

	dapi, err := NewDeltaAPI(deltaApi, deltaAuthToken)
	if err != nil {
		log.Fatalf("could not connect to delta api: %s", err)
	} else {
		log.Debugf("successfully connected to db at %s\n", deltaApi)
	}

	return &DeltaDM{
		DAPI: dapi,
		DB:   db,
	}
}