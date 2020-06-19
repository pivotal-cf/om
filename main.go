package main

import (
	"github.com/pivotal-cf/om/cmd"
	_ "github.com/pivotal-cf/om/download_clients"
	"log"
	"os"
)

var version = "unknown"

var applySleepDurationString  = "10s"


func main() {
	err := cmd.Main(os.Stdout, os.Stderr, version, applySleepDurationString, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
