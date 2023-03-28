package cmd

import (
	"github.com/urfave/cli/v2"
)

func ReplicationCmd() []*cli.Command {
	var ddmApi string
	var deltaAuthToken string

	// add a command to run API node
	var replicationCmds []*cli.Command
	replicationCmd := &cli.Command{
		Name:  "wallet",
		Usage: "Interact with DDM wallets",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "ddm-api-info",
				Usage:       "DDM API connection info",
				EnvVars:     []string{"DDM_API_INFO"},
				DefaultText: "http://localhost:1314",
				Value:       "http://localhost:1314",
				Destination: &ddmApi,
			},
			&cli.StringFlag{
				Name:        "delta-auth",
				Usage:       "delta auth token",
				EnvVars:     []string{"DELTA_AUTH"},
				Required:    true,
				Destination: &deltaAuthToken,
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "import",
				Usage: "import a wallet to DDM",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "file",
						Usage: "path to wallet file",
					},
					&cli.StringFlag{
						Name:  "hex",
						Usage: "wallet data in hex (lotus export) format",
					},
				},
				Action: func(c *cli.Context) error {

					return nil
				},
			},
		},
	}

	replicationCmds = append(replicationCmds, replicationCmd)

	return replicationCmds
}
