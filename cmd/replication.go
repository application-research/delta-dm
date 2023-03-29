package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/application-research/delta-dm/api"
	"github.com/urfave/cli/v2"
)

func ReplicationCmd() []*cli.Command {
	var num uint
	var provider string
	var dataset string

	// add a command to run API node
	var replicationCmds []*cli.Command
	replicationCmd := &cli.Command{
		Name:  "replication",
		Usage: "Dataset Replications",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create dataset replications with provider",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:        "num",
						Usage:       "number of deals to make",
						Destination: &num,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "provider",
						Usage:       "storage provider to make deals with",
						Destination: &provider,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "dataset",
						Usage:       "dataset to replicate",
						Destination: &dataset,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					body := api.PostReplicationBody{
						NumDeals: &num,
						Provider: provider,
					}

					if dataset != "" {
						body.Dataset = &dataset
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unabel to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest("POST", "/api/v1/replications/", b)
					if err != nil {
						return fmt.Errorf("unable to make request %s", err)
					}
					defer closer()

					log.Printf("replication response: %s", string(res))

					return nil
				},
			},
		},
	}

	replicationCmds = append(replicationCmds, replicationCmd)

	return replicationCmds
}
