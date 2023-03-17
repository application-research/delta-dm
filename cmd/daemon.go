package cmd

import (
	"github.com/application-research/delta-dm/api"
	"github.com/application-research/delta-dm/core"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func DaemonCmd() []*cli.Command {
	var debug bool = false
	var dbConnStr string
	var deltaApi string
	var deltaAuthToken string
	var authUrl string

	// add a command to run API node
	var daemonCommands []*cli.Command
	daemonCmd := &cli.Command{
		Name:  "daemon",
		Usage: "A server-side application for orchestrating dataset dealmaking to Filecoin SPs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "db",
				Usage:       "connection string for postgres db",
				EnvVars:     []string{"DB_NAME"},
				DefaultText: "delta-dm.db",
				Value:       "delta-dm.db",
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
				Usage:       "delta auth token, used for service connection (database sync)",
				EnvVars:     []string{"DELTA_AUTH"},
				Required:    true,
				Destination: &deltaAuthToken,
			},
			&cli.StringFlag{
				Name:        "auth-url",
				Usage:       "auth server URL. defaults to use official Estuary auth service. specify to use custom auth server",
				EnvVars:     []string{"AUTH_URL"},
				DefaultText: "https://auth.estuary.tech",
				Value:       "https://auth.estuary.tech",
				Destination: &authUrl,
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

			dldm := core.NewDeltaDM(dbConnStr, deltaApi, deltaAuthToken, authUrl, debug)
			dldm.WatchReplications()
			api.InitializeEchoRouterConfig(dldm)
			api.LoopForever()

			return nil
		},
	}

	daemonCommands = append(daemonCommands, daemonCmd)

	return daemonCommands
}
