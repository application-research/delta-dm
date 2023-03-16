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
						Name:  "file",
						Usage: "path to wallet file",
					},
					&cli.StringFlag{
						Name:  "hex",
						Usage: "wallet data in hex (lotus export) format",
					},
				},
				Action: func(c *cli.Context) error {
					cp, err := NewCmdProcessor(ddmApi, deltaAuthToken)

					if err != nil {
						return err
					}

					walletPath := c.String("file")
					walletHex := c.String("hex")

					if walletPath == "" && walletHex == "" {
						return fmt.Errorf("must provide either wallet file, or hex wallet export")
					}

					if walletPath != "" && walletHex != "" {
						return fmt.Errorf("please provide either wallet file or hex, not both")
					}

					var walletBytes []byte

					if walletPath != "" {
						var walletData WalletJSON
						walletFile, err := ioutil.ReadFile(walletPath)
						if err != nil {
							return fmt.Errorf("failed to open wallet file: %s", err)
						}

						// Verify parsing
						err = json.Unmarshal(walletFile, &walletData)
						if err != nil {
							return fmt.Errorf("failed to parse wallet file: %s", err)
						}

						if walletData.Type == "" || walletData.PrivateKey == "" {
							return fmt.Errorf("wallet data must contain Type and PrivateKey")
						}

						walletBytes, err = json.Marshal(walletData)
						if err != nil {
							return fmt.Errorf("failed to prepare wallet json: %s", err)
						}
					} else {
						var walletHexRequest WalletHex = WalletHex{HexKey: walletHex}
						// Hex import
						walletBytes, err = json.Marshal(walletHexRequest)
						if err != nil {
							return fmt.Errorf("failed to parse wallet hex: %s", err)
						}
					}

					url := "/api/v1/wallets"
					if walletHex != "" {
						url += "?hex=true"
					}

					res, closer, err := cp.ddmPostRequest(url, walletBytes)
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
type WalletHex struct {
	HexKey string `json:"hex_key"`
}
