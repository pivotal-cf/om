package main

import (
	"errors"
	"log"
	"os"

	"github.com/pivotal-cf/om/cmd"
	"github.com/pivotal-cf/om/commands"
	_ "github.com/pivotal-cf/om/download_clients"
)

var version = "unknown"

var applySleepDurationString = "10s"

func main() {
	os.Args = append(os.Args, "--target=34.68.121.89")
	os.Args = append(os.Args, "--username=admin")
	os.Args = append(os.Args, "--password=J768f2AwJpP9FA3yWXL2")
	os.Args = append(os.Args, "--skip-ssl-validation")
	os.Args = append(os.Args, "expiring-licenses")

	err := cmd.Main(os.Stdout, os.Stderr, version, applySleepDurationString, os.Args)
	if err != nil {
		if errors.Is(err, commands.ErrBoshDiffChangesExist) {
			log.Print(err)
			os.Exit(2)
		}
		log.Fatal(err)
	}
}
