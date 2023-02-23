package main

import (
	"os"

	cmd "github.com/application-research/delta-dm/cmd"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

var (
	log = logging.Logger("api")
)

func main() {
	var commands []*cli.Command

	// commands
	commands = append(commands, cmd.DaemonCmd()...)
	// commands = append(commands, cmd.CommpCmd()...)
	app := &cli.App{
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
