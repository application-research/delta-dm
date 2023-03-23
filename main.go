package main

import (
	"fmt"
	"os"

	cmd "github.com/application-research/delta-dm/cmd"
	"github.com/application-research/delta-dm/core"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

var (
	log = logging.Logger("api")
)

var Commit string
var Version string

func main() {
	var commands []*cli.Command

	di := core.DeploymentInfo{
		Commit:  Commit,
		Version: Version,
	}

	versionCmd := cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Action: func(c *cli.Context) error {
			fmt.Printf("ddm version : %s+git.%s\n", di.Version, di.Commit)
			return nil
		},
	}

	// commands
	commands = append(commands, &versionCmd)
	commands = append(commands, cmd.DaemonCmd(di)...)
	commands = append(commands, cmd.WalletCmd()...)
	app := &cli.App{
		Commands: commands,
		Usage:    "An application to facilitate dataset dealmaking with storage providers",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
