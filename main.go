package main

import (
	"log"
	"os"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/flags"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/progress"
)

var version = "unknown"

func main() {
	logger := log.New(os.Stdout, "", 0)

	var global struct {
		Version           bool   `short:"v" long:"version"             description:"prints the om release version"                        default:"false"`
		Help              bool   `short:"h" long:"help"                description:"prints this usage information"                        default:"false"`
		Target            string `short:"t" long:"target"              description:"location of the Ops Manager VM"`
		Username          string `short:"u" long:"username"            description:"admin username for the Ops Manager VM"`
		Password          string `short:"p" long:"password"            description:"admin password for the Ops Manager VM"`
		SkipSSLValidation bool   `short:"k" long:"skip-ssl-validation" description:"skip ssl certificate validation during http requests" default:"false"`
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
		command, args = args[0], args[1:]
	}

	if global.Version {
		command = "version"
	}

	if global.Help {
		command = "help"
	}

	unauthenticatedClient := network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation)

	authedClient, err := network.NewOAuthClient(global.Target, global.Username, global.Password, global.SkipSSLValidation)
	if err != nil {
		logger.Fatal(err)
	}

	setupService := api.NewSetupService(unauthenticatedClient)
	uploadStemcellService := api.NewUploadStemcellService(authedClient, progress.NewBar())

	commandSet := commands.Set{}
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["version"] = commands.NewVersion(version, os.Stdout)
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(setupService, logger)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(formcontent.NewForm("stemcell[file]"), uploadStemcellService, logger)

	err = commandSet.Execute(command, args)
	if err != nil {
		logger.Fatal(err)
	}
}
