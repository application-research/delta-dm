package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

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
				Name:  "import",
				Usage: "import a wallet to DDM",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "json",
						Usage: "wallet data in json format",
					},
					&cli.StringFlag{
						Name:  "file",
						Usage: "path to wallet file",
					},
					&cli.StringFlag{
						Name:  "dataset",
						Usage: "dataset name to associate wallet with",
					},
				},
				Action: func(c *cli.Context) error {
					cp, err := NewCmdProcessor(ddmApi, deltaAuthToken)

					if err != nil {
						return err
					}

					walletJson := c.String("json")
					walletPath := c.String("file")
					dataset := c.String("dataset")

					if walletJson == "" && walletPath == "" {
						return fmt.Errorf("must provide either json or file")
					}

					if walletJson != "" && walletPath != "" {
						return fmt.Errorf("please provide either wallet JSON or file path, not both")
					}

					var walletData WalletJSON

					if walletPath != "" {
						walletFile, err := ioutil.ReadFile(walletPath)
						if err != nil {
							return fmt.Errorf("failed to open wallet file: %s", err)
						}

						// Verify parsing
						err = json.Unmarshal(walletFile, &walletData)
						if err != nil {
							return fmt.Errorf("failed to parse wallet file: %s", err)
						}

					} else {
						err = json.Unmarshal([]byte(walletJson), &walletData)
						if err != nil {
							return fmt.Errorf("failed to parse wallet json: %s", err)
						}
					}

					if walletData.Type == "" || walletData.PrivateKey == "" {
						return fmt.Errorf("wallet data must contain Type and PrivateKey")
					}

					// Ignoring error here as we know it's been unmarshaled by this point
					wb, _ := json.Marshal(walletData)

					url := "/api/v1/wallets"
					if dataset != "" {
						url += "?dataset=" + dataset
					}

					res, closer, err := cp.ddmPostRequest(url, wb)
					if err != nil {
						return err
					}
					defer closer()

					log.Printf("Wallet import response: %s", string(res))
					return nil
				},
			},
		},
	}

	walletCmds = append(walletCmds, walletCmd)

	return walletCmds
}

type WalletJSON struct {
	Type       string `json:"Type"`
	PrivateKey string `json:"PrivateKey"`
}
