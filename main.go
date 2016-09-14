package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "unknown"

func main() {
	var printVersion bool

	set := flag.NewFlagSet("default", flag.ContinueOnError)
	set.SetOutput(os.Stdout)
	set.BoolVar(&printVersion, "v", false, "prints the version")
	set.BoolVar(&printVersion, "version", false, "prints the version")
	err := set.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	if len(set.Args()) > 0 {
		switch set.Args()[0] {
		case "version":
			printVersion = true
		default:
			set.PrintDefaults()
			os.Exit(1)
		}
	}

	if printVersion {
		fmt.Println(version)
	}
}
