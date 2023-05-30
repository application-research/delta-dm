package core

import (
	db "github.com/application-research/delta-dm/db"
	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"
)

var (
	log = logging.Logger("router")
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

	dbi, err := db.OpenDatabase(dbConnStr, debug)
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

	as := NewAuthServer(authServerUrl, authToken)

	return &DeltaDM{
		DAPI:       dapi,
		DB:         dbi,
		AS:         as,
		Info:       di,
		DryRunMode: dryRun,
	}
}
