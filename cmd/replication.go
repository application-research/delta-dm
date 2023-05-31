package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/application-research/delta-dm/api"
	"github.com/urfave/cli/v2"
)

func ReplicationCmd() []*cli.Command {
	var num uint
	var provider string
	var datasetID uint
	var delayStartDays uint64

	var replicationCmds []*cli.Command
	replicationCmd := &cli.Command{
		Name:  "replication",
		Usage: "Dataset Replications Commands",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create dataset replications with provider",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:        "num",
						Aliases:     []string{"n"},
						Usage:       "number of deals to make",
						Destination: &num,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "provider",
						Aliases:     []string{"p"},
						Usage:       "storage provider to make deals with",
						Destination: &provider,
						Required:    true,
					},
					&cli.UintFlag{
						Name:        "dataset",
						Aliases:     []string{"d"},
						Usage:       "dataset id to replicate",
						Destination: &datasetID,
					},
					&cli.Uint64Flag{
						Name:        "delay-start",
						Usage:       "number of days to delay start of deal",
						Destination: &delayStartDays,
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

					if datasetID != 0 {
						body.DatasetID = &datasetID
					}

					if delayStartDays != 0 {
						body.DelayStartDays = &delayStartDays
					}

					b, err := json.Marshal(body)
					if err != nil {
						return fmt.Errorf("unable to construct request body %s", err)
					}

					res, closer, err := cmd.MakeRequest(http.MethodPost, "/api/v1/replications", b)
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

	replicationCmds = append(replicationCmds, replicationCmd)

	return replicationCmds
}
