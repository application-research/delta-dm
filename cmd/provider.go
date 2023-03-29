package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/urfave/cli/v2"
)

func ProviderCmd() []*cli.Command {
	var spId string
	var spName string
	var allowSelfService bool

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

					res, closer, err := cmd.MakeRequest("POST", "/api/v1/providers", b)
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
						Aliases:     []string{"id"},
						Usage:       "storage provider id to modify (i.e. f012345)",
						Destination: &spId,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "name",
						Usage:       "update friendly name of storage provider",
						Destination: &spName,
					},
					&cli.BoolFlag{
						Name:        "allow-selfserve",
						Aliases:     []string{"id"},
						Usage:       "update enable self-service for provider",
						Destination: &allowSelfService,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := core.Provider{
						ActorName:        spName,
						AllowSelfService: allowSelfService,
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest("POST", "/api/v1/providers/"+spId, b)
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

					res, closer, err := cmd.MakeRequest("GET", "/api/v1/providers/", nil)
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
