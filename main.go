package main

import (
	"errors"
	"log"
	"os"

	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"

	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/flags"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/progress"
)

var version = "unknown"

const applySleepSeconds = 10

func main() {
	liveWriter := uilive.New()

	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	var global struct {
		Version           bool   `short:"v" long:"version"             description:"prints the om release version"                        default:"false"`
		Help              bool   `short:"h" long:"help"                description:"prints this usage information"                        default:"false"`
		Target            string `short:"t" long:"target"              description:"location of the Ops Manager VM"`
		Username          string `short:"u" long:"username"            description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
		Password          string `short:"p" long:"password"            description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
		SkipSSLValidation bool   `short:"k" long:"skip-ssl-validation" description:"skip ssl certificate validation during http requests" default:"false"`
		RequestTimeout    int    `short:"r" long:"request-timeout"     description:"timeout in seconds for HTTP requests to Ops Manager" default:"1800"`
	}

	args, err := flags.Parse(&global, os.Args[1:])
	if err != nil {
		stdout.Fatal(err)
	}

	globalFlagsUsage, err := flags.Usage(global)
	if err != nil {
		stdout.Fatal(err)
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

	if global.Target == "" && command != "help" && command != "version" {
		stdout.Fatal(errors.New("error: target flag is required. Run `om help` for more info."))
	}

	requestTimeout := time.Duration(global.RequestTimeout) * time.Second

	unauthenticatedClient := network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation, requestTimeout)

	authedClient, err := network.NewOAuthClient(global.Target, global.Username, global.Password, global.SkipSSLValidation, false, requestTimeout)
	if err != nil {
		stdout.Fatal(err)
	}

	authedCookieClient, err := network.NewOAuthClient(global.Target, global.Username, global.Password, global.SkipSSLValidation, true, requestTimeout)
	if err != nil {
		stdout.Fatal(err)
	}

	setupService := api.NewSetupService(unauthenticatedClient)
	uploadStemcellService := api.NewUploadStemcellService(authedClient, progress.NewBar())
	stagedProductsService := api.NewStagedProductsService(authedClient)
	availableProductsService := api.NewAvailableProductsService(authedClient, progress.NewBar(), liveWriter)
	diagnosticService := api.NewDiagnosticService(authedClient)
	importInstallationService := api.NewInstallationAssetService(unauthenticatedClient, progress.NewBar(), liveWriter)
	exportInstallationService := api.NewInstallationAssetService(authedClient, progress.NewBar(), liveWriter)
	deleteInstallationService := api.NewInstallationAssetService(authedClient, nil, nil)
	installationsService := api.NewInstallationsService(authedClient)
	logWriter := commands.NewLogWriter(os.Stdout)
	tableWriter := tablewriter.NewWriter(os.Stdout)
	requestService := api.NewRequestService(authedClient)
	jobsService := api.NewJobsService(authedClient)
	boshService := api.NewBoshFormService(authedCookieClient)
	dashboardService := api.NewDashboardService(authedCookieClient)

	form, err := formcontent.NewForm()
	if err != nil {
		stdout.Fatal(err)
	}

	extractor := extractor.ProductUnzipper{}

	commandSet := commands.Set{}
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["version"] = commands.NewVersion(version, os.Stdout)
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(setupService, stdout)
	commandSet["configure-bosh"] = commands.NewConfigureBosh(boshService, stdout)
	commandSet["revert-staged-changes"] = commands.NewRevertStagedChanges(dashboardService, stdout)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(form, uploadStemcellService, diagnosticService, stdout)
	commandSet["upload-product"] = commands.NewUploadProduct(form, extractor, availableProductsService, stdout)
	commandSet["delete-unused-products"] = commands.NewDeleteUnusedProducts(availableProductsService, stdout)
	commandSet["stage-product"] = commands.NewStageProduct(stagedProductsService, availableProductsService, diagnosticService, stdout)
	commandSet["configure-product"] = commands.NewConfigureProduct(stagedProductsService, jobsService, stdout)
	commandSet["export-installation"] = commands.NewExportInstallation(exportInstallationService, stdout)
	commandSet["import-installation"] = commands.NewImportInstallation(form, importInstallationService, setupService, stdout)
	commandSet["delete-installation"] = commands.NewDeleteInstallation(deleteInstallationService, installationsService, logWriter, stdout, applySleepSeconds)
	commandSet["apply-changes"] = commands.NewApplyChanges(installationsService, logWriter, stdout, applySleepSeconds)
	commandSet["curl"] = commands.NewCurl(requestService, stdout, stderr)
	commandSet["available-products"] = commands.NewAvailableProducts(availableProductsService, tableWriter, stdout)

	err = commandSet.Execute(command, args)
	if err != nil {
		stdout.Fatal(err)
	}
}
