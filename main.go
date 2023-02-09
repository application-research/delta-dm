package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
)

type DeltaLDM struct {
	deal_length   uint64
	piece_size    uint64
	car_directory string

	db *gorm.DB
}

func main() {
	// May need Lotus
	var car_directory string
	var debug bool = false
	var dbConnStr string

	app := &cli.App{
		Name:      "Delta Large Dataset Manager",
		Usage:     "A server-side application for orchestrating large dataset dealmaking to Filecoin SPs",
		UsageText: "./terror --dir /car/dir",
		Version:   "0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "directory where CAR files are located",
				Required:    true,
				Destination: &car_directory,
			},
			&cli.StringFlag{
				Name:        "db",
				Usage:       "connection string for postgres db",
				EnvVars:     []string{"DB"},
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

			db, err := setupDatabase(dbConnStr)
			if err != nil {
				log.Fatalf("could not connect to db: %s", err)
			}

			dldm := DeltaLDM{
				car_directory: car_directory,
				deal_length:   1514800,
				piece_size:    34359738368,
				db:            db,
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
	e := echo.New()

	api := e.Group("/api")

	api.POST("/deal", t.HandlePostDeal)
}

func setupDatabase(dbConnStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbConnStr), &gorm.Config{})

	return db, err
}
