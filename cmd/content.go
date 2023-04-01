package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/application-research/delta-dm/core"
	"github.com/urfave/cli/v2"
)

func ContentCmd() []*cli.Command {
	var datasetName string

	// add a command to run API node
	var contentCmds []*cli.Command
	contentCmd := &cli.Command{
		Name:  "content",
		Usage: "Contents (CAR files) Commands",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add content to a dataset",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "dataset",
						Aliases:     []string{"d"},
						Usage:       "dataset name (slug)",
						Destination: &datasetName,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}

					inputCnt := c.Args().First()

					if inputCnt == "" {
						return fmt.Errorf("must provide content")
					}

					var cnt []core.Content

					if err := json.Unmarshal([]byte(inputCnt), &cnt); err != nil {
						return fmt.Errorf("couldn't parse content: %s", err)
					}

					res, closer, err := cmd.MakeRequest("POST", fmt.Sprintf("/contents/%s", datasetName), []byte(inputCnt))

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

	contentCmds = append(contentCmds, contentCmd)

	return contentCmds
}
