package main

import (
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"

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

	type globalFlags struct {
		ConfigFile        string `                           short:"tc" long:"target-config-file"                  description:"YAML file specifying om global parameters"`
		ClientID          string `yaml:"client-id"           short:"c"  long:"client-id"                           description:"Client ID for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_ID)"`
		ClientSecret      string `yaml:"client-secret"       short:"s"  long:"client-secret"                       description:"Client Secret for the Ops Manager VM (not required for unauthenticated commands, $OM_CLIENT_SECRET)"`
		Format            string `yaml:"format"              short:"f"  long:"format"              default:"table" description:"Format to print as (options: table,json)"`
		Help              bool   `                           short:"h"  long:"help"                default:"false" description:"prints this usage information"`
		Password          string `yaml:"password"            short:"p"  long:"password"                            description:"admin password for the Ops Manager VM (not required for unauthenticated commands, $OM_PASSWORD)"`
		RequestTimeout    int    `yaml:"request-timeout"     short:"r"  long:"request-timeout"     default:"1800"  description:"timeout in seconds for HTTP requests to Ops Manager"`
		SkipSSLValidation bool   `yaml:"skip-ssl-validation" short:"k"  long:"skip-ssl-validation" default:"false" description:"skip ssl certificate validation during http requests"`
		Target            string `yaml:"target"              short:"t"  long:"target"                              description:"location of the Ops Manager VM"`
		Trace             bool   `yaml:"trace"               short:"tr" long:"trace"                               description:"prints HTTP requests and response payloads"`
		Username          string `yaml:"username"            short:"u"  long:"username"                            description:"admin username for the Ops Manager VM (not required for unauthenticated commands, $OM_USERNAME)"`
		Version           bool   `                           short:"v"  long:"version"             default:"false" description:"prints the om release version"`
	}

	var configFlags globalFlags

	args, err := jhanda.Parse(&configFlags, os.Args[1:])
	if err != nil {
		stdout.Fatal(err)
	}

	globalFlagsUsage, err := jhanda.PrintUsage(configFlags)
	if err != nil {
		stdout.Fatal(err)
	}

	if configFlags.ConfigFile != "" {
		fileContents, err := ioutil.ReadFile(configFlags.ConfigFile)
		if err != nil {
			stdout.Fatal(err)
		}

		var fileFlags globalFlags
		err = yaml.Unmarshal(fileContents, &fileFlags)
		if err != nil {
			stdout.Fatal(err)
		}

		if configFlags.ClientID == "" {
			configFlags.ClientID = fileFlags.ClientID
		}
		if configFlags.ClientSecret == "" {
			configFlags.ClientSecret = fileFlags.ClientSecret
		}
		if configFlags.Format == "" {
			configFlags.Format = fileFlags.Format
		}
		if configFlags.Password == "" {
			configFlags.Password = fileFlags.Password
		}
		if configFlags.RequestTimeout == 1800 {
			configFlags.RequestTimeout = fileFlags.RequestTimeout
		}
		if !configFlags.SkipSSLValidation {
			configFlags.SkipSSLValidation = fileFlags.SkipSSLValidation
		}
		if configFlags.Target == "" {
			configFlags.Target = fileFlags.Target
		}
		if !configFlags.Trace {
			configFlags.Trace = fileFlags.Trace
		}
		if configFlags.Username == "" {
			configFlags.Username = fileFlags.Username
		}
	}

	var command string
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}

	if configFlags.Version {
		command = "version"
	}

	if configFlags.Help {
		command = "help"
	}

	if command == "" {
		command = "help"
	}

	if configFlags.Username == "" {
		configFlags.Username = os.Getenv("OM_USERNAME")
	}

	if configFlags.Password == "" {
		configFlags.Password = os.Getenv("OM_PASSWORD")
	}

	if configFlags.ClientID == "" {
		configFlags.ClientID = os.Getenv("OM_CLIENT_ID")
	}

	if configFlags.ClientSecret == "" {
		configFlags.ClientSecret = os.Getenv("OM_CLIENT_SECRET")
	}

	requestTimeout := time.Duration(configFlags.RequestTimeout) * time.Second

	var unauthenticatedClient, authedClient, authedCookieClient, unauthenticatedProgressClient, authedProgressClient httpClient
	unauthenticatedClient = network.NewUnauthenticatedClient(configFlags.Target, configFlags.SkipSSLValidation, requestTimeout)
	authedClient, err = network.NewOAuthClient(configFlags.Target, configFlags.Username, configFlags.Password, configFlags.ClientID, configFlags.ClientSecret, configFlags.SkipSSLValidation, false, requestTimeout)
	if err != nil {
		stdout.Fatal(err)
	}
	authedCookieClient, err = network.NewOAuthClient(configFlags.Target, configFlags.Username, configFlags.Password, configFlags.ClientID, configFlags.ClientSecret, configFlags.SkipSSLValidation, true, requestTimeout)
	if err != nil {
		stdout.Fatal(err)
	}

	liveWriter := uilive.New()
	liveWriter.Out = os.Stderr
	unauthenticatedProgressClient = network.NewProgressClient(unauthenticatedClient, progress.NewBar(), liveWriter)
	authedProgressClient = network.NewProgressClient(authedClient, progress.NewBar(), liveWriter)

	if configFlags.Trace {
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
	switch configFlags.Format {
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
