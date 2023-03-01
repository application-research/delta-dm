package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func WalletCmd() []*cli.Command {
	var ddmApi string
	var deltaAuthToken string

	// add a command to run API node
	var walletCmds []*cli.Command
	walletCmd := &cli.Command{
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
				Name:  "add",
				Usage: "Add a wallet to DDM",
				Action: func(c *cli.Context) error {
					_, err := NewCmdProcessor(ddmApi, deltaAuthToken)

					if err != nil {
						return err
					}

					fmt.Println("success!")

					return nil
				},
			},
		},
	}

	walletCmds = append(walletCmds, walletCmd)

	return walletCmds
}
