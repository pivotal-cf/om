package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/gosuri/uilive"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/go-pivnet/logshim"
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

var applySleepDurationString = "10s"

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type options struct {
	DecryptionPassphrase string `yaml:"decryption-passphrase" short:"d" long:"decryption-passphrase" env:"OM_DECRYPTION_PASSPHRASE" description:"Passphrase to decrypt the installation if the Ops Manager VM has been rebooted (optional for most commands)"`
	ClientID             string `yaml:"client-id"            short:"c"  long:"client-id"           env:"OM_CLIENT_ID"                     description:"Client ID for the Ops Manager VM (not required for unauthenticated commands)"`
	ClientSecret         string `yaml:"client-secret"        short:"s"  long:"client-secret"       env:"OM_CLIENT_SECRET"                 description:"Client Secret for the Ops Manager VM (not required for unauthenticated commands)"`
	Help                 bool   `                            short:"h"  long:"help"                                       default:"false" description:"prints this usage information"`
	Password             string `yaml:"password"             short:"p"  long:"password"            env:"OM_PASSWORD"                      description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
	ConnectTimeout       int    `yaml:"connect-timeout"      short:"o"  long:"connect-timeout"                            default:"5"     description:"timeout in seconds to make TCP connections"`
	RequestTimeout       int    `yaml:"request-timeout"      short:"r"  long:"request-timeout"                            default:"1800"  description:"timeout in seconds for HTTP requests to Ops Manager"`
	SkipSSLValidation    bool   `yaml:"skip-ssl-validation"  short:"k"  long:"skip-ssl-validation"                        default:"false" description:"skip ssl certificate validation during http requests"`
	Target               string `yaml:"target"               short:"t"  long:"target"              env:"OM_TARGET"                        description:"location of the Ops Manager VM"`
	Trace                bool   `yaml:"trace"                short:"tr" long:"trace"                                                      description:"prints HTTP requests and response payloads"`
	Username             string `yaml:"username"             short:"u"  long:"username"            env:"OM_USERNAME"                      description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
	Env                  string `                            short:"e"  long:"env"                                                        description:"env file with login credentials"`
	Version              bool   `                            short:"v"  long:"version"                                    default:"false" description:"prints the om release version"`
}

func main() {
	applySleepDuration, _ := time.ParseDuration(applySleepDurationString)

	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	var global options

	args, err := jhanda.Parse(&global, os.Args[1:])
	if err != nil {
		stderr.Fatal(err)
	}

	err = setEnvFileProperties(&global)
	if err != nil {
		stderr.Fatal(err)
	}

	globalFlagsUsage, err := jhanda.PrintUsage(global)
	if err != nil {
		stderr.Fatal(err)
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

	requestTimeout := time.Duration(global.RequestTimeout) * time.Second
	connectTimeout := time.Duration(global.ConnectTimeout) * time.Second

	var unauthenticatedClient, authedClient, authedCookieClient, unauthenticatedProgressClient, authedProgressClient httpClient
	unauthenticatedClient = network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation, requestTimeout, connectTimeout)
	authedClient, err = network.NewOAuthClient(global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, false, requestTimeout, connectTimeout)

	if global.DecryptionPassphrase != "" {
		authedClient = network.NewDecryptClient(authedClient, unauthenticatedClient, global.DecryptionPassphrase, os.Stderr)
	}

	if err != nil {
		stderr.Fatal(err)
	}
	authedCookieClient, err = network.NewOAuthClient(global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, true, requestTimeout, connectTimeout)
	if err != nil {
		stderr.Fatal(err)
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
	pivnetLogWriter := logshim.NewLogShim(stdout, stdout, global.Trace)

	form := formcontent.NewForm()

	metadataExtractor := extractor.MetadataExtractor{}

	pivnetFactory := commands.DefaultPivnetFactory

	presenter := presenters.NewPresenter(presenters.NewTablePresenter(tableWriter), presenters.NewJSONPresenter(os.Stdout))

	commandSet := jhanda.CommandSet{}
	commandSet["activate-certificate-authority"] = commands.NewActivateCertificateAuthority(api, stdout)
	commandSet["apply-changes"] = commands.NewApplyChanges(api, api, logWriter, stdout, applySleepDuration)
	commandSet["available-products"] = commands.NewAvailableProducts(api, presenter, stdout)
	commandSet["certificate-authorities"] = commands.NewCertificateAuthorities(api, presenter)
	commandSet["certificate-authority"] = commands.NewCertificateAuthority(api, presenter, stdout)
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(api, stdout)
	commandSet["configure-director"] = commands.NewConfigureDirector(os.Environ, api, stdout)
	commandSet["configure-product"] = commands.NewConfigureProduct(os.Environ, api, stdout)
	commandSet["configure-saml-authentication"] = commands.NewConfigureSAMLAuthentication(api, stdout)
	commandSet["config-template"] = commands.NewConfigTemplate(metadataExtractor, stdout)
	commandSet["create-certificate-authority"] = commands.NewCreateCertificateAuthority(api, presenter)
	commandSet["create-vm-extension"] = commands.NewCreateVMExtension(os.Environ, api, stdout)
	commandSet["credential-references"] = commands.NewCredentialReferences(api, presenter, stdout)
	commandSet["credentials"] = commands.NewCredentials(api, presenter, stdout)
	commandSet["curl"] = commands.NewCurl(api, stdout, stderr)
	commandSet["delete-certificate-authority"] = commands.NewDeleteCertificateAuthority(api, stdout)
	commandSet["delete-installation"] = commands.NewDeleteInstallation(api, logWriter, stdout, applySleepDuration)
	commandSet["delete-product"] = commands.NewDeleteProduct(api)
	commandSet["delete-unused-products"] = commands.NewDeleteUnusedProducts(api, stdout)
	commandSet["deployed-manifest"] = commands.NewDeployedManifest(api, stdout)
	commandSet["deployed-products"] = commands.NewDeployedProducts(presenter, api)
	commandSet["download-product"] = commands.NewDownloadProduct(os.Environ, pivnetLogWriter, os.Stdout, pivnetFactory)
	commandSet["errands"] = commands.NewErrands(presenter, api)
	commandSet["export-installation"] = commands.NewExportInstallation(api, stderr)
	commandSet["generate-certificate"] = commands.NewGenerateCertificate(api, stdout)
	commandSet["generate-certificate-authority"] = commands.NewGenerateCertificateAuthority(api, presenter)
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["import-installation"] = commands.NewImportInstallation(form, api, global.DecryptionPassphrase, stdout)
	commandSet["installation-log"] = commands.NewInstallationLog(api, stdout)
	commandSet["installations"] = commands.NewInstallations(api, presenter)
	commandSet["interpolate"] = commands.NewInterpolate(os.Environ, stdout)
	commandSet["pending-changes"] = commands.NewPendingChanges(presenter, api)
	commandSet["regenerate-certificates"] = commands.NewRegenerateCertificates(api, stdout)
	commandSet["revert-staged-changes"] = commands.NewRevertStagedChanges(ui, stdout)
	commandSet["staged-config"] = commands.NewStagedConfig(api, stdout)
	commandSet["staged-director-config"] = commands.NewStagedDirectorConfig(api, stdout)
	commandSet["stage-product"] = commands.NewStageProduct(api, stdout)
	commandSet["staged-manifest"] = commands.NewStagedManifest(api, stdout)
	commandSet["staged-products"] = commands.NewStagedProducts(presenter, api)
	commandSet["tile-metadata"] = commands.NewTileMetadata(stdout)
	commandSet["unstage-product"] = commands.NewUnstageProduct(api, stdout)
	commandSet["upload-product"] = commands.NewUploadProduct(form, metadataExtractor, api, stdout)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(form, api, stdout)
	commandSet["version"] = commands.NewVersion(version, os.Stdout)

	err = commandSet.Execute(command, args)
	if err != nil {
		stderr.Fatal(err)
	}
}

func setEnvFileProperties(global *options) error {
	if global.Env == "" {
		return nil
	}

	var opts options
	file, err := os.Open(global.Env)
	if err != nil {
		return fmt.Errorf("env file does not exist: %s", err)
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read env file: %s", err)
	}

	err = yaml.Unmarshal(contents, &opts)
	if err != nil {
		return fmt.Errorf("could not parse env file: %s", err)
	}

	if global.ClientID == "" {
		global.ClientID = opts.ClientID
	}
	if global.ClientSecret == "" {
		global.ClientSecret = opts.ClientSecret
	}
	if global.Password == "" {
		global.Password = opts.Password
	}
	if global.ConnectTimeout == 5 && opts.ConnectTimeout != 0 {
		global.ConnectTimeout = opts.ConnectTimeout
	}
	if global.RequestTimeout == 1800 && opts.RequestTimeout != 0 {
		global.RequestTimeout = opts.RequestTimeout
	}
	if global.SkipSSLValidation == false {
		global.SkipSSLValidation = opts.SkipSSLValidation
	}
	if global.Target == "" {
		global.Target = opts.Target
	}
	if global.Trace == false {
		global.Trace = opts.Trace
	}
	if global.Username == "" {
		global.Username = opts.Username
	}
	if global.DecryptionPassphrase == "" {
		global.DecryptionPassphrase = opts.DecryptionPassphrase
	}

	return nil
}
