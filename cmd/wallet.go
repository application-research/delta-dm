package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/urfave/cli/v2"
)

func WalletCmd() []*cli.Command {

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
				Usage:     "Delete a wallet in DDM",
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
