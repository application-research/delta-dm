package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/api"
	"github.com/application-research/delta-dm/core"
	"github.com/urfave/cli/v2"
)

func ProviderCmd() []*cli.Command {
	var spId string
	var spName string
	var allowSelfService string
	var allowedDatasets cli.StringSlice

	// add a command to run API node
	var providerCmds []*cli.Command
	providerCmd := &cli.Command{
		Name:  "provider",
		Usage: "Storage Provider Commands",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add storage provider",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "id",
						Usage:       "storage provider id to add (i.e. f012345)",
						Destination: &spId,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "name",
						Usage:       "friendly name of storage provider",
						Destination: &spName,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := core.Provider{
						ActorID: spId,
					}

					if spName != "" {
						body.ActorName = spName
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPost, "/api/v1/providers", b)
					if err != nil {
						return fmt.Errorf("unable to make request %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))

					return nil
				},
			},
			{
				Name:  "modify",
				Usage: "modify existing storage provider",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "id",
						Usage:       "storage provider id to modify (i.e. f012345)",
						Destination: &spId,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "update friendly name of storage provider",
						Destination: &spName,
					},
					&cli.StringFlag{
						Name:        "allow-selfserve",
						Aliases:     []string{"ss"},
						Usage:       "enable self-service for provider (on|off)",
						Destination: &allowSelfService,
					},
					&cli.StringSliceFlag{
						Name:        "allowed-datasets",
						Usage:       "datasets the provider is permitted to replicate (comma separated list)",
						Destination: &allowedDatasets,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					if allowSelfService != "" {
						if allowSelfService != "on" && allowSelfService != "off" {
							return fmt.Errorf("allow-selfserve must be 'on' or 'off'")
						}
					}

					body := api.ProviderPutBody{
						ActorName:        spName,
						AllowSelfService: allowSelfService,
						AllowedDatasets:  allowedDatasets.Value(),
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPut, "/api/v1/providers/"+spId, b)
					if err != nil {
						return fmt.Errorf("unable to make request %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))

					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list storage providers",
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					res, closer, err := cmd.MakeRequest(http.MethodGet, "/api/v1/providers/", nil)
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

	providerCmds = append(providerCmds, providerCmd)

	return providerCmds
}
