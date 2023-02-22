package main

import (
	"os"

	api "github.com/application-research/delta-ldm/api"
	core "github.com/application-research/delta-ldm/core"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type DeltaLDM struct {
	dapi *core.DeltaAPI
	db   *gorm.DB
}

func main() {
	var debug bool = false
	var dbConnStr string
	var deltaApi string
	var deltaAuthToken string

	app := &cli.App{
		Name:      "Delta Large Dataset Manager",
		Usage:     "A server-side application for orchestrating large dataset dealmaking to Filecoin SPs",
		UsageText: "./delta-ldm",
		Version:   "0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "db",
				Usage:       "connection string for postgres db",
				EnvVars:     []string{"DB_NAME"},
				Required:    true,
				Destination: &dbConnStr,
			},
			&cli.StringFlag{
				Name:        "delta-api",
				Usage:       "endpoint for delta api",
				DefaultText: "http://localhost:1414",
				Value:       "http://localhost:1414",
				EnvVars:     []string{"DELTA_API"},
				Destination: &deltaApi,
			},
			&cli.StringFlag{
				Name:        "delta-auth",
				Usage:       "delta auth token",
				EnvVars:     []string{"DELTA_AUTH"},
				Required:    true,
				Destination: &deltaAuthToken,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "set to enable debug logging output",
				Destination: &debug,
			},
		},

		Action: func(cctx *cli.Context) error {
			if debug {
				log.SetLevel(log.DebugLevel)
			}

			db, err := core.OpenDatabase(dbConnStr)
			if err != nil {
				log.Fatalf("could not connect to db: %s", err)
			} else {
				log.Debugf("successfully connected to delta api at %s\n", deltaApi)
			}

			dapi, err := core.NewDeltaAPI(deltaApi, deltaAuthToken)
			if err != nil {
				log.Fatalf("could not connect to delta api: %s", err)
			} else {
				log.Debugf("successfully connected to db at %s\n", deltaApi)
			}

			dldm := DeltaLDM{
				dapi: dapi,
				db:   db,
			}
			dldm.serveAPI()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (t *DeltaLDM) serveAPI() {
	api.InitializeEchoRouterConfig(t.db)
	api.LoopForever()
}
