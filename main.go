package main

import (
	"log"
	"os"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/flags"
)

var version = "unknown"

func main() {
	logger := log.New(os.Stdout, "", 0)

	var global struct {
		Version bool `short:"v" long:"version" default:"false" description:"prints the om release version"`
		Help    bool `short:"h" long:"help"    default:"false" description:"prints this usage information"`
	}
	args, err := flags.Parse(&global, os.Args[1:])
	if err != nil {
		logger.Fatal(err)
	}

	globalFlagsUsage, err := flags.Usage(global)
	if err != nil {
		logger.Fatal(err)
	}

	var command string
	if len(args) > 0 {
		command = args[0]
	}

	if global.Version {
		command = "version"
	}

	if global.Help {
		command = "help"
	}

	versionCommand := commands.NewVersion(version, os.Stdout)
	helpCommand := commands.NewHelp(os.Stdout, globalFlagsUsage, versionCommand)

	commandSet := commands.Set{
		"version": versionCommand,
		"help":    helpCommand,
	}

	err = commandSet.Execute(command)
	if err != nil {
		logger.Fatal(err)
	}
}
