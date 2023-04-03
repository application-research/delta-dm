package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

func ContentCmd() []*cli.Command {
	var datasetName string

	// add a command to run API node
	var contentCmds []*cli.Command
	contentCmd := &cli.Command{
		Name:  "content",
		Usage: "Content (CAR files) Commands",
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
					&cli.StringFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "content in ddm json format",
					},
					&cli.StringFlag{
						Name:    "csv",
						Aliases: []string{"c"},
						Usage:   "filename of content in csv format",
					},
					&cli.StringFlag{
						Name:    "singularity",
						Aliases: []string{"s"},
						Usage:   "content in singularity json export format",
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}
					jsonData := c.String("json")
					csvFilename := c.String("csv")
					singularityData := c.String("singularity")

					if jsonData == "" && csvFilename == "" && singularityData == "" {
						return fmt.Errorf("must either json, singularity or csv flag")
					}

					var body []byte
					url := "/contents/" + datasetName

					if jsonData != "" {
						body = []byte(jsonData)
					} else if csvFilename != "" {
						csvFile, err := ioutil.ReadFile(csvFilename)
						if err != nil {
							return fmt.Errorf("failed to open csv file: %s", err)
						}
						body = csvFile
						url += "?import_type=csv"
					} else if singularityData != "" {
						body = []byte(singularityData)
						url += "?import_type=singularity"
					}

					res, closer, err := cmd.MakeRequest("POST", fmt.Sprintf("/contents/%s", datasetName), body)

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
