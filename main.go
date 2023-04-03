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

	// commands
	commands = append(commands, cmd.DaemonCmd(di)...)
	commands = append(commands, cmd.WalletCmd()...)
	commands = append(commands, cmd.ReplicationCmd()...)
	commands = append(commands, cmd.ProviderCmd()...)
	commands = append(commands, cmd.DatasetCmd()...)
	commands = append(commands, cmd.ContentCmd()...)

	app := &cli.App{
		Commands: commands,
		Usage:    "An application to facilitate dataset dealmaking with storage providers",
		Version:  fmt.Sprintf("%s+git.%s\n", di.Version, di.Commit),
		Flags:    cmd.CLIConnectFlags,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
