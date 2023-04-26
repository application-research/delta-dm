package core

import (
	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"
)

type DeploymentInfo struct {
	Commit  string `json:"commit"`
	Version string `json:"version"`
}
type DeltaDM struct {
	DAPI       *DeltaAPI
	DB         *gorm.DB
	AS         *AuthServer
	Info       DeploymentInfo
	DryRunMode bool
}

func NewDeltaDM(dbConnStr string, deltaApi string, authToken string, authServerUrl string, di DeploymentInfo, debug bool, dryRun bool) *DeltaDM {
	if debug {
		logging.SetDebugLogging()
	}

	db, err := OpenDatabase(dbConnStr, debug)
	if err != nil {
		log.Fatalf("could not connect to db: %s", err)
	} else {
		log.Debugf("successfully connected to delta db at %s\n", deltaApi)
	}

	dapi, err := NewDeltaAPI(deltaApi, authToken)
	if err != nil {
		log.Fatalf("could not connect to delta api: %s", err)
	} else {
		log.Debugf("successfully connected to api at %s\n", deltaApi)
	}

	as := NewAuthServer(authServerUrl)

	return &DeltaDM{
		DAPI:       dapi,
		DB:         db,
		AS:         as,
		Info:       di,
		DryRunMode: dryRun,
	}
}
