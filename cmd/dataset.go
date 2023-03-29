package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/core"
	"github.com/application-research/delta-dm/util"
	"github.com/urfave/cli/v2"
)

func DatasetCmd() []*cli.Command {
	var datasetName string
	var replicationQuota uint64
	var dealDuration uint64

	// add a command to run API node
	var datasetCmds []*cli.Command
	datasetCmd := &cli.Command{
		Name:  "dataset",
		Usage: "Dataset Commands",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add dataset",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "dataset name (slug)",
						Destination: &datasetName,
						Required:    true,
					},
					&cli.Uint64Flag{
						Name:        "replicaion-quota",
						Aliases:     []string{"q"},
						Usage:       "replication quota - how many times the dataset may be replicated",
						DefaultText: "6",
						Value:       6,
						Destination: &replicationQuota,
					},
					&cli.Uint64Flag{
						Name:        "duration",
						Aliases:     []string{"d"},
						Usage:       "deal duration - how long (in days) should deals for this dataset last",
						DefaultText: "540",
						Value:       540,
						Destination: &dealDuration,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					if !util.ValidateDatasetName(datasetName) {
						return fmt.Errorf("invalid dataset name. must contain only lowercase letters, numbers and hyphens. must begin and end with a letter. must not contain consecutive hyphens")
					}

					body := core.Dataset{
						Name:             datasetName,
						ReplicationQuota: replicationQuota,
						DealDuration:     dealDuration,
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPost, "/api/v1/datasets", b)
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
				Usage: "list datasets",
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					res, closer, err := cmd.MakeRequest(http.MethodGet, "/api/v1/datasets", nil)
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

	datasetCmds = append(datasetCmds, datasetCmd)

	return datasetCmds
}
