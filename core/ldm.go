package core

import (
	"gorm.io/gorm"
)

type DeltaLDM struct {
	DAPI *DeltaAPI
	DB   *gorm.DB
}

func NewDeltaLDM(dbConnStr string, deltaApi string, deltaAuthToken string) *DeltaLDM {
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

	return &DeltaLDM{
		DAPI: dapi,
		DB:   db,
	}
}
