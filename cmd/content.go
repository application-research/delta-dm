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
						Usage:   "filename of content in ddm json format",
					},
					&cli.StringFlag{
						Name:    "csv",
						Aliases: []string{"c"},
						Usage:   "filename of content in csv format",
					},
					&cli.StringFlag{
						Name:    "singularity",
						Aliases: []string{"s"},
						Usage:   "filename of content in singularity json export format",
					},
				},
				Action: func(c *cli.Context) error {
					cmd, err := NewCmdProcessor(c)
					if err != nil {
						return err
					}
					jsonFilename := c.String("json")
					csvFilename := c.String("csv")
					singularityDataFilename := c.String("singularity")

					if jsonFilename == "" && csvFilename == "" && singularityDataFilename == "" {
						return fmt.Errorf("must either json, singularity or csv flag")
					}

					var body []byte
					url := "/api/v1/contents/" + datasetName

					if jsonFilename != "" {
						jsonFile, err := ioutil.ReadFile(jsonFilename)
						if err != nil {
							return fmt.Errorf("failed to open json file: %s", err)
						}
						body = jsonFile
					} else if csvFilename != "" {
						csvFile, err := ioutil.ReadFile(csvFilename)
						if err != nil {
							return fmt.Errorf("failed to open csv file: %s", err)
						}
						body = csvFile
						url += "?import_type=csv"
					} else if singularityDataFilename != "" {
						singularityFile, err := ioutil.ReadFile(singularityDataFilename)
						if err != nil {
							return fmt.Errorf("failed to open singularity json file: %s", err)
						}
						body = singularityFile
						url += "?import_type=singularity"
					}

					res, closer, err := cmd.MakeRequest("POST", url, body)

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
