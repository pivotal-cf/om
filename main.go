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

	global := flags.NewGlobal()
	args, err := global.Parse(os.Args[1:])
	if err != nil {
		logger.Fatal(err)
	}

	var command string
	if len(args) > 0 {
		command = args[0]
	}

	if global.Version.Value {
		command = "version"
	}

	if global.Help.Value {
		command = "help"
	}

	versionCommand := commands.NewVersion(version, os.Stdout)
	helpCommand := commands.NewHelp([]commands.Helper{
		global.Version,
		global.Help,
	}, []commands.Helper{
		versionCommand,
	}, os.Stdout)

	commandSet := commands.Set{
		"version": versionCommand,
		"help":    helpCommand,
	}

	err = commandSet.Execute(command)
	if err != nil {
		logger.Fatal(err)
	}
}
