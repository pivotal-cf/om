package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pivotal-cf/om/cmd"
	"github.com/pivotal-cf/om/commands"
	_ "github.com/pivotal-cf/om/download_clients"
)

var version = "unknown"

var applySleepDurationString = "10s"

func main() {
	err := cmd.Main(os.Stdout, os.Stderr, version, applySleepDurationString, os.Args)
	if err != nil {
		if errors.Is(err, commands.ErrBoshDiffChangesExist) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
