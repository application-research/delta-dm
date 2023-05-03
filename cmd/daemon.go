package cmd

import (
	"fmt"

	"github.com/application-research/delta-dm/api"
	"github.com/application-research/delta-dm/core"
	"github.com/application-research/delta-dm/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func DaemonCmd(di core.DeploymentInfo) []*cli.Command {
	var debug bool = false
	var dryRun bool = false
	var dbConnStr string
	var deltaApi string
	var deltaAuthToken string
	var authServer string
	var port uint

	var daemonCommands []*cli.Command
	daemonCmd := &cli.Command{
		Name:  "daemon",
		Usage: "A server-side application for orchestrating dataset dealmaking to Filecoin SPs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "db",
				Usage:       "connection string for postgres db or sqlite db filename",
				EnvVars:     []string{"DB_DSN"},
				DefaultText: "delta-dm.db",
				Value:       "delta-dm.db",
				Destination: &dbConnStr,
			},
			&cli.UintFlag{
				Name:        "port",
				Usage:       "port that delta-dm will run on",
				EnvVars:     []string{"DELTA_DM_PORT"},
				DefaultText: "1415",
				Value:       1415,
				Destination: &port,
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
				Usage:       "delta auth token, used for service connection (database sync)",
				EnvVars:     []string{"DELTA_AUTH"},
				Required:    true,
				Destination: &deltaAuthToken,
			},
			&cli.StringFlag{
				Name:        "auth-server",
				Usage:       "auth server URL. defaults to official Estuary auth service. specify to use custom auth server",
				EnvVars:     []string{"AUTH_SERVER"},
				DefaultText: "https://auth.estuary.tech",
				Value:       "https://auth.estuary.tech",
				Destination: &authServer,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "set to enable debug logging output",
				Destination: &debug,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Hidden:      true,
				Usage:       "don't actually make deals (for development and testing)",
				Destination: &dryRun,
			},
		},

		Action: func(cctx *cli.Context) error {
			if debug {
				log.SetLevel(log.DebugLevel)
			}

			fmt.Println(util.Green + "Delta DM - By Protocol Labs - Outercore Engineering" + util.Reset)

			fmt.Println(util.Blue + "Starting DDM daemon..." + util.Reset)

			if dryRun {
				fmt.Println(util.Yellow + "Running in dry-run mode. No deals will be made." + util.Reset)
			}

			dldm := core.NewDeltaDM(dbConnStr, deltaApi, deltaAuthToken, authServer, di, debug, dryRun)
			dldm.WatchReplications()
			api.InitializeEchoRouterConfig(dldm, port)
			api.LoopForever()

			return nil
		},
	}

	daemonCommands = append(daemonCommands, daemonCmd)

	return daemonCommands
}
