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

const applySleepSeconds = 1

func main() {
	logger := log.New(os.Stdout, "", 0)

	var global struct {
		Version           bool   `short:"v" long:"version"             description:"prints the om release version"                        default:"false"`
		Help              bool   `short:"h" long:"help"                description:"prints this usage information"                        default:"false"`
		Target            string `short:"t" long:"target"              description:"location of the Ops Manager VM"`
		Username          string `short:"u" long:"username"            description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
		Password          string `short:"p" long:"password"            description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
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

	if command == "" {
		command = "help"
	}

	unauthenticatedClient := network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation)

	authedClient, err := network.NewOAuthClient(global.Target, global.Username, global.Password, global.SkipSSLValidation)
	if err != nil {
		logger.Fatal(err)
	}

	setupService := api.NewSetupService(unauthenticatedClient)
	uploadStemcellService := api.NewUploadStemcellService(authedClient, progress.NewBar())
	productService := api.NewProductService(authedClient, progress.NewBar())
	diagnosticService := api.NewDiagnosticService(authedClient)
	importInstallationService := api.NewInstallationAssetService(unauthenticatedClient, progress.NewBar())
	exportInstallationService := api.NewInstallationAssetService(authedClient, progress.NewBar())
	installationsService := api.NewInstallationsService(authedClient)
	logWriter := commands.NewLogWriter(os.Stdout)

	form, err := formcontent.NewForm()
	if err != nil {
		logger.Fatal(err)
	}

	commandSet := commands.Set{}
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["version"] = commands.NewVersion(version, os.Stdout)
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(setupService, logger)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(form, uploadStemcellService, diagnosticService, logger)
	commandSet["stage-product"] = commands.NewStageProduct(productService, logger)
	commandSet["upload-product"] = commands.NewUploadProduct(form, productService, logger)
	commandSet["export-installation"] = commands.NewExportInstallation(exportInstallationService, logger)
	commandSet["import-installation"] = commands.NewImportInstallation(form, importInstallationService, setupService, logger)
	commandSet["apply-changes"] = commands.NewApplyChanges(installationsService, logWriter, logger, applySleepSeconds)

	err = commandSet.Execute(command, args)
	if err != nil {
		logger.Fatal(err)
	}
}
