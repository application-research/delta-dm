package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/application-research/delta-dm/api"
	"github.com/urfave/cli/v2"
)

func WalletCmd() []*cli.Command {
	var dataset string
	var walletAddress string

	// add a command to run API node
	var walletCmds []*cli.Command
	walletCmd := &cli.Command{
		Name:  "wallet",
		Usage: "Interact with DDM wallets",
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
					cp, err := NewCmdProcessor(c)
					if err != nil {
						return fmt.Errorf("failed to connect to ddm node: %s", err)
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

					res, closer, err := cp.MakeRequest(http.MethodPost, url, walletBytes)
					if err != nil {
						return fmt.Errorf("ddm request invalid: %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))
					return nil
				},
			},
			{
				Name:      "delete",
				Usage:     "delete a wallet",
				UsageText: "delta-dm wallet delete [wallet address]",
				Action: func(c *cli.Context) error {
					cp, err := NewCmdProcessor(c)
					if err != nil {
						return fmt.Errorf("failed to connect to ddm node: %s", err)
					}

					w := c.Args().First()

					if w == "" {
						return fmt.Errorf("please provide a wallet address")
					}

					res, closer, err := cp.MakeRequest(http.MethodDelete, "/api/v1/wallets/"+w, nil)
					if err != nil {
						return fmt.Errorf("ddm request invalid: %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))
					return nil
				},
			},
			{
				Name:      "associate",
				Usage:     "associate wallet with dataset",
				UsageText: "delta-dm wallet associate [wallet address]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "datasets",
						Usage:       "dataset names to associate with wallet (comma separated)",
						Destination: &dataset,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "address",
						Usage:       "wallet address to associate",
						Destination: &walletAddress,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					cp, err := NewCmdProcessor(c)
					if err != nil {
						return fmt.Errorf("failed to connect to ddm node: %s", err)
					}

					datasets := strings.Split(dataset, ",")

					if len(datasets) < 1 {
						return fmt.Errorf("please provide at least one dataset name")
					}

					awb := api.AssociateWalletBody{
						Address:  walletAddress,
						Datasets: datasets,
					}

					b, err := json.Marshal(awb)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cp.MakeRequest(http.MethodPost, "/api/v1/wallets/associate", b)
					if err != nil {
						return fmt.Errorf("ddm request invalid: %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list wallets",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dataset",
						Usage:       "filter wallets by dataset",
						Destination: &dataset,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					url := "/api/v1/wallets"

					if dataset != "" {
						url += "?dataset=" + dataset
					}

					res, closer, err := cmd.MakeRequest(http.MethodGet, url, nil)
					if err != nil {
						return fmt.Errorf("unable to make request %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))

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
