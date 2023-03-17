package core

import (
	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"
)

type DeltaDM struct {
	DAPI *DeltaAPI
	DB   *gorm.DB
	AS   *AuthServer
}

func NewDeltaDM(dbConnStr string, deltaApi string, authToken string, authServerUrl string, debug bool) *DeltaDM {
	if debug {
		logging.SetDebugLogging()
	}

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

	as := NewAuthServer(authServerUrl)

	return &DeltaDM{
		DAPI: dapi,
		DB:   db,
		AS:   as,
	}
}
