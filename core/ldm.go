package core

import (
	"gorm.io/gorm"
)

type DeltaDM struct {
	DAPI *DeltaAPI
	DB   *gorm.DB
}

func NewDeltaDM(dbConnStr string, deltaApi string, authToken string, debug bool) *DeltaDM {
	db, err := OpenDatabase(dbConnStr, debug)
	if err != nil {
		log.Fatalf("could not connect to db: %s", err)
	} else {
		log.Debugf("successfully connected to delta api at %s\n", deltaApi)
	}

	dapi, err := NewDeltaAPI(deltaApi, authToken)
	if err != nil {
		log.Fatalf("could not connect to delta api: %s", err)
	} else {
		log.Debugf("successfully connected to db at %s\n", deltaApi)
	}

	err = RunReconciliation(db, dapi)
	if err != nil {
		log.Fatalf("could not reconcile: %s", err)
	}

	return &DeltaDM{
		DAPI: dapi,
		DB:   db,
	}
}
