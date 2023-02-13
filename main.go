package main

import (
	"os"

	api "github.com/application-research/delta-ldm/api"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type DeltaLDM struct {
	db *gorm.DB
}

func main() {
	var debug bool = false
	var dbConnStr string

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

			db, err := OpenDatabase(dbConnStr)
			if err != nil {
				log.Fatalf("could not connect to db: %s", err)
			}

			dldm := DeltaLDM{
				db: db,
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
	api.InitializeEchoRouterConfig()
	api.LoopForever()
}
