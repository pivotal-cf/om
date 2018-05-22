package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"

	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/presenters"
	"github.com/pivotal-cf/om/progress"
	"github.com/pivotal-cf/om/ui"
)

var version = "unknown"

const applySleepSeconds = 10

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func main() {
	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	var global struct {
		ClientID          string `short:"c"  long:"client-id"                           description:"Client ID for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_ID)"`
		ClientSecret      string `short:"s"  long:"client-secret"                       description:"Client Secret for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_SECRET)"`
		Format            string `short:"f"  long:"format"              default:"table" description:"Format to print as (options: table,json)"`
		Help              bool   `short:"h"  long:"help"                default:"false" description:"prints this usage information"`
		Password          string `short:"p"  long:"password"                            description:"admin password for the Ops Manager VM (not required for unauthenticated commands, $OM_PASSWORD)"`
		ConnectTimeout    int    `short:"o"  long:"connect-timeout"     default:"5"     description:"timeout in seconds to make TCP connections"`
		RequestTimeout    int    `short:"r"  long:"request-timeout"     default:"1800"  description:"timeout in seconds for HTTP requests to Ops Manager"`
		SkipSSLValidation bool   `short:"k"  long:"skip-ssl-validation" default:"false" description:"skip ssl certificate validation during http requests"`
		Target            string `short:"t"  long:"target"                              description:"location of the Ops Manager VM"`
		Trace             bool   `short:"tr" long:"trace"                               description:"prints HTTP requests and response payloads"`
		Username          string `short:"u"  long:"username"                            description:"admin username for the Ops Manager VM (not required for unauthenticated commands, $OM_USERNAME)"`
		Version           bool   `short:"v"  long:"version"             default:"false" description:"prints the om release version"`
	}

	args, err := jhanda.Parse(&global, os.Args[1:])
	if err != nil {
		stdout.Fatal(err)
	}

	globalFlagsUsage, err := jhanda.PrintUsage(global)
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

	if global.Username == "" {
		global.Username = os.Getenv("OM_USERNAME")
	}

	if global.Password == "" {
		global.Password = os.Getenv("OM_PASSWORD")
	}

	if global.ClientID == "" {
		global.ClientID = os.Getenv("OM_CLIENT_ID")
	}

	if global.ClientSecret == "" {
		global.ClientSecret = os.Getenv("OM_CLIENT_SECRET")
	}

	requestTimeout := time.Duration(global.RequestTimeout) * time.Second
	connectTimeout := time.Duration(global.ConnectTimeout) * time.Second

	var unauthenticatedClient, authedClient, authedCookieClient, unauthenticatedProgressClient, authedProgressClient httpClient
	unauthenticatedClient = network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation, requestTimeout, connectTimeout)
	authedClient, err = network.NewOAuthClient(global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, false, requestTimeout, connectTimeout)
	if err != nil {
		stdout.Fatal(err)
	}
	authedCookieClient, err = network.NewOAuthClient(global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, true, requestTimeout, connectTimeout)
	if err != nil {
		stdout.Fatal(err)
	}

	liveWriter := uilive.New()
	liveWriter.Out = os.Stderr
	unauthenticatedProgressClient = network.NewProgressClient(unauthenticatedClient, progress.NewBar(), liveWriter)
	authedProgressClient = network.NewProgressClient(authedClient, progress.NewBar(), liveWriter)

	if global.Trace {
		unauthenticatedClient = network.NewTraceClient(unauthenticatedClient, os.Stderr)
		unauthenticatedProgressClient = network.NewTraceClient(unauthenticatedProgressClient, os.Stderr)
		authedClient = network.NewTraceClient(authedClient, os.Stderr)
		authedCookieClient = network.NewTraceClient(authedCookieClient, os.Stderr)
		authedProgressClient = network.NewTraceClient(authedProgressClient, os.Stderr)
	}

	api := api.New(api.ApiInput{
		Client:                 authedClient,
		UnauthedClient:         unauthenticatedClient,
		ProgressClient:         authedProgressClient,
		UnauthedProgressClient: unauthenticatedProgressClient,
		Logger:                 stderr,
	})
	ui := ui.New(ui.UiInput{
		Client: authedCookieClient,
	})
	logWriter := commands.NewLogWriter(os.Stdout)
	tableWriter := tablewriter.NewWriter(os.Stdout)

	form, err := formcontent.NewForm()
	if err != nil {
		stdout.Fatal(err)
	}

	metadataExtractor := extractor.MetadataExtractor{}

	var presenter presenters.Presenter
	switch global.Format {
	case "table":
		presenter = presenters.NewTablePresenter(tableWriter)
	case "json":
		presenter = presenters.NewJSONPresenter(os.Stdout)
	default:
		stdout.Fatal("Format not supported")
	}

	commandSet := jhanda.CommandSet{}
	commandSet["activate-certificate-authority"] = commands.NewActivateCertificateAuthority(api, stdout)
	commandSet["apply-changes"] = commands.NewApplyChanges(api, logWriter, stdout, applySleepSeconds)
	commandSet["available-products"] = commands.NewAvailableProducts(api, presenter, stdout)
	commandSet["certificate-authorities"] = commands.NewCertificateAuthorities(api, presenter)
	commandSet["certificate-authority"] = commands.NewCertificateAuthority(api, presenter, stdout)
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(api, stdout)
	commandSet["configure-bosh"] = commands.NewConfigureBosh(ui, api, stdout, stderr)
	commandSet["configure-director"] = commands.NewConfigureDirector(api, stdout)
	commandSet["configure-product"] = commands.NewConfigureProduct(api, stdout)
	commandSet["configure-saml-authentication"] = commands.NewConfigureSAMLAuthentication(api, stdout)
	commandSet["config-template"] = commands.NewConfigTemplate(metadataExtractor, stdout)
	commandSet["create-certificate-authority"] = commands.NewCreateCertificateAuthority(api, presenter)
	commandSet["create-vm-extension"] = commands.NewCreateVMExtension(api, stdout)
	commandSet["credential-references"] = commands.NewCredentialReferences(api, presenter, stdout)
	commandSet["credentials"] = commands.NewCredentials(api, presenter, stdout)
	commandSet["curl"] = commands.NewCurl(api, stdout, stderr)
	commandSet["delete-certificate-authority"] = commands.NewDeleteCertificateAuthority(api, stdout)
	commandSet["delete-installation"] = commands.NewDeleteInstallation(api, logWriter, stdout, applySleepSeconds)
	commandSet["delete-product"] = commands.NewDeleteProduct(api)
	commandSet["delete-unused-products"] = commands.NewDeleteUnusedProducts(api, stdout)
	commandSet["deployed-manifest"] = commands.NewDeployedManifest(api, stdout)
	commandSet["deployed-products"] = commands.NewDeployedProducts(presenter, api)
	commandSet["errands"] = commands.NewErrands(presenter, api)
	commandSet["export-installation"] = commands.NewExportInstallation(api, stderr)
	commandSet["generate-certificate"] = commands.NewGenerateCertificate(api, stdout)
	commandSet["generate-certificate-authority"] = commands.NewGenerateCertificateAuthority(api, presenter)
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["import-installation"] = commands.NewImportInstallation(form, api, stdout)
	commandSet["installation-log"] = commands.NewInstallationLog(api, stdout)
	commandSet["installations"] = commands.NewInstallations(api, presenter)
	commandSet["interpolate"] = commands.NewInterpolate(stdout)
	commandSet["pending-changes"] = commands.NewPendingChanges(presenter, api)
	commandSet["regenerate-certificates"] = commands.NewRegenerateCertificates(api, stdout)
	commandSet["revert-staged-changes"] = commands.NewRevertStagedChanges(ui, stdout)
	commandSet["set-errand-state"] = commands.NewSetErrandState(api)
	commandSet["staged-config"] = commands.NewStagedConfig(api, stdout)
	commandSet["stage-product"] = commands.NewStageProduct(api, stdout)
	commandSet["staged-manifest"] = commands.NewStagedManifest(api, stdout)
	commandSet["staged-products"] = commands.NewStagedProducts(presenter, api)
	commandSet["unstage-product"] = commands.NewUnstageProduct(api, stdout)
	commandSet["upload-product"] = commands.NewUploadProduct(form, metadataExtractor, api, stdout)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(form, api, stdout)
	commandSet["version"] = commands.NewVersion(version, os.Stdout)

	err = commandSet.Execute(command, args)
	if err != nil {
		stderr.Fatal(err)
	}
}
