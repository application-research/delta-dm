package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/core"
	"github.com/urfave/cli/v2"
)

func ReplicationProfilesCmd() []*cli.Command {
	var spId string
	var datasetId uint
	var unsealed bool
	var indexed bool

	// add a command to run API node
	var providerCmds []*cli.Command
	providerCmd := &cli.Command{
		Name:    "replication-profile",
		Aliases: []string{"rp"},
		Usage:   "Replication Profile Commands",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add replication profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "spid",
						Usage:       "storage provider id",
						Destination: &spId,
						Required:    true,
					},
					&cli.UintFlag{
						Name:        "dataset",
						Usage:       "dataset id",
						Destination: &datasetId,
						Required:    true,
					},
					&cli.BoolFlag{
						Name:        "indexed",
						Usage:       "announce deals to indexer",
						Destination: &indexed,
					},
					&cli.BoolFlag{
						Name:        "unsealed",
						Usage:       "keep unsealed copy",
						Destination: &unsealed,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := core.ReplicationProfile{
						ProviderActorID: spId,
						DatasetID:       datasetId,
						Unsealed:        unsealed,
						Indexed:         indexed,
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPost, "/api/v1/replication-profiles/", b)
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
				Usage: "modify replication profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "spid",
						Usage:       "storage provider id",
						Destination: &spId,
						Required:    true,
					},
					&cli.UintFlag{
						Name:        "dataset",
						Usage:       "dataset id",
						Destination: &datasetId,
						Required:    true,
					},
					&cli.BoolFlag{
						Name:        "indexed",
						Usage:       "announce deals to indexer",
						Destination: &indexed,
					},
					&cli.BoolFlag{
						Name:        "unsealed",
						Usage:       "keep unsealed copy",
						Destination: &unsealed,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := core.ReplicationProfile{
						ProviderActorID: spId,
						DatasetID:       datasetId,
						Unsealed:        unsealed,
						Indexed:         indexed,
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPut, "/api/v1/replication-profiles/", b)
					if err != nil {
						return fmt.Errorf("unable to make request %s", err)
					}
					defer closer()

					fmt.Printf("%s", string(res))

					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "delete replication profile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "spid",
						Usage:       "storage provider id",
						Destination: &spId,
						Required:    true,
					},
					&cli.UintFlag{
						Name:        "dataset",
						Usage:       "dataset id",
						Destination: &datasetId,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := core.ReplicationProfile{
						ProviderActorID: spId,
						DatasetID:       datasetId,
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodDelete, "/api/v1/replication-profiles/", b)
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
				Usage: "list replication profiles",
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					res, closer, err := cmd.MakeRequest(http.MethodGet, "/api/v1/replication-profiles/", nil)
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
