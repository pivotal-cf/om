package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/interpolate"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/presenters"
	"github.com/pivotal-cf/om/renderers"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type options struct {
	CACert               string `yaml:"ca-cert" long:"ca-cert" env:"OM_CA_CERT" description:"OpsManager CA certificate path or value"`
	ClientID             string `yaml:"client-id"             short:"c"  long:"client-id"             env:"OM_CLIENT_ID"                           description:"Client ID for the Ops Manager VM (not required for unauthenticated commands)"`
	ClientSecret         string `yaml:"client-secret"         short:"s"  long:"client-secret"         env:"OM_CLIENT_SECRET"                       description:"Client Secret for the Ops Manager VM (not required for unauthenticated commands)"`
	ConnectTimeout       int    `yaml:"connect-timeout"       short:"o"  long:"connect-timeout"       env:"OM_CONNECT_TIMEOUT"     default:"10"    description:"timeout in seconds to make TCP connections"`
	DecryptionPassphrase string `yaml:"decryption-passphrase" short:"d"  long:"decryption-passphrase" env:"OM_DECRYPTION_PASSPHRASE"             description:"Passphrase to decrypt the installation if the Ops Manager VM has been rebooted (optional for most commands)"`
	Env                  string `                             short:"e"  long:"env"                                                              description:"env file with login credentials"`
	Help                 bool   `                             short:"h"  long:"help"                                             default:"false" description:"prints this usage information"`
	Password             string `yaml:"password"              short:"p"  long:"password"              env:"OM_PASSWORD"                            description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
	RequestTimeout       int    `yaml:"request-timeout"       short:"r"  long:"request-timeout"       env:"OM_REQUEST_TIMEOUT"     default:"1800"  description:"timeout in seconds for HTTP requests to Ops Manager"`
	SkipSSLValidation    bool   `yaml:"skip-ssl-validation"   short:"k"  long:"skip-ssl-validation"   env:"OM_SKIP_SSL_VALIDATION" default:"false" description:"skip ssl certificate validation during http requests"`
	Target               string `yaml:"target"                short:"t"  long:"target"                env:"OM_TARGET"                              description:"location of the Ops Manager VM"`
	Trace                bool   `yaml:"trace"                 short:"tr" long:"trace"                 env:"OM_TRACE"                               description:"prints HTTP requests and response payloads"`
	Username             string `yaml:"username"              short:"u"  long:"username"              env:"OM_USERNAME"                            description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
	VarsEnv              string `                                                                     env:"OM_VARS_ENV"                            description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
	Version              bool   `                             short:"v"  long:"version"                                          default:"false" description:"prints the om release version"`
}

func Main(sout io.Writer, serr io.Writer, version string, applySleepDurationString string, args []string) error {
	applySleepDuration, _ := time.ParseDuration(applySleepDurationString)

	stdout := log.New(sout, "", 0)
	stderr := log.New(serr, "", 0)

	var global options

	args, err := jhanda.Parse(&global, args[1:])
	if err != nil {
		return err
	}

	err = setEnvFileProperties(&global)
	if err != nil {
		return err
	}

	globalFlagsUsage, err := jhanda.PrintUsage(global)
	if err != nil {
		return err
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

	var unauthenticatedClient, authedClient, unauthenticatedProgressClient, authedProgressClient httpClient
	unauthenticatedClient, err = network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation, global.CACert, connectTimeout, requestTimeout)
	if err != nil {
		return err
	}

	authedClient, err = network.NewOAuthClient(global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, global.CACert, connectTimeout, requestTimeout)

	if err != nil {
		return err
	}

	if global.DecryptionPassphrase != "" {
		authedClient = network.NewDecryptClient(authedClient, unauthenticatedClient, global.DecryptionPassphrase, os.Stderr)
	}

	unauthenticatedProgressClient = network.NewProgressClient(unauthenticatedClient, os.Stderr)
	authedProgressClient = network.NewProgressClient(authedClient, os.Stderr)

	if global.Trace {
		unauthenticatedClient = network.NewTraceClient(unauthenticatedClient, os.Stderr)
		unauthenticatedProgressClient = network.NewTraceClient(unauthenticatedProgressClient, os.Stderr)
		authedClient = network.NewTraceClient(authedClient, os.Stderr)
		authedProgressClient = network.NewTraceClient(authedProgressClient, os.Stderr)
	}

	api := api.New(api.ApiInput{
		Client:                 authedClient,
		UnauthedClient:         unauthenticatedClient,
		ProgressClient:         authedProgressClient,
		UnauthedProgressClient: unauthenticatedProgressClient,
		Logger:                 stderr,
	})

	logWriter := commands.NewLogWriter(os.Stdout)
	tableWriter := tablewriter.NewWriter(os.Stdout)

	form := formcontent.NewForm()

	metadataExtractor := extractor.NewMetadataExtractor()

	presenter := presenters.NewPresenter(presenters.NewTablePresenter(tableWriter), presenters.NewJSONPresenter(os.Stdout))
	envRendererFactory := renderers.NewFactory(renderers.NewEnvGetter())

	commandSet := jhanda.CommandSet{}
	commandSet["activate-certificate-authority"] = commands.NewActivateCertificateAuthority(api, stdout)
	commandSet["apply-changes"] = commands.NewApplyChanges(api, api, logWriter, stdout, applySleepDuration)
	commandSet["assign-multi-stemcell"] = commands.NewAssignMultiStemcell(api, stdout)
	commandSet["assign-stemcell"] = commands.NewAssignStemcell(api, stdout)
	commandSet["available-products"] = commands.NewAvailableProducts(api, presenter, stdout)
	commandSet["bosh-diff"] = commands.NewBoshDiff(api, stdout)
	commandSet["bosh-env"] = commands.NewBoshEnvironment(api, stdout, global.Target, envRendererFactory)
	commandSet["certificate-authorities"] = commands.NewCertificateAuthorities(api, presenter)
	commandSet["certificate-authority"] = commands.NewCertificateAuthority(api, presenter, stdout)
	commandSet["config-template"] = commands.NewConfigTemplate(commands.DefaultProvider())
	commandSet["configure-authentication"] = commands.NewConfigureAuthentication(os.Environ, api, stdout)
	commandSet["configure-director"] = commands.NewConfigureDirector(os.Environ, api, stdout)
	commandSet["configure-ldap-authentication"] = commands.NewConfigureLDAPAuthentication(os.Environ, api, stdout)
	commandSet["configure-opsman"] = commands.NewConfigureOpsman(os.Environ, api, stderr)
	commandSet["configure-product"] = commands.NewConfigureProduct(os.Environ, api, global.Target, stdout)
	commandSet["configure-saml-authentication"] = commands.NewConfigureSAMLAuthentication(os.Environ, api, stdout)
	commandSet["create-certificate-authority"] = commands.NewCreateCertificateAuthority(api, presenter)
	commandSet["create-vm-extension"] = commands.NewCreateVMExtension(os.Environ, api, stdout)
	commandSet["credential-references"] = commands.NewCredentialReferences(api, presenter, stdout)
	commandSet["credentials"] = commands.NewCredentials(api, presenter, stdout)
	commandSet["curl"] = commands.NewCurl(api, stdout, stderr)
	commandSet["delete-certificate-authority"] = commands.NewDeleteCertificateAuthority(api, stdout)
	commandSet["delete-installation"] = commands.NewDeleteInstallation(api, logWriter, stdout, os.Stdin, applySleepDuration)
	commandSet["delete-product"] = commands.NewDeleteProduct(api)
	commandSet["delete-ssl-certificate"] = commands.NewDeleteSSLCertificate(api, stdout)
	commandSet["delete-unused-products"] = commands.NewDeleteUnusedProducts(api, stdout)
	commandSet["deployed-manifest"] = commands.NewDeployedManifest(api, stdout)
	commandSet["deployed-products"] = commands.NewDeployedProducts(presenter, api)
	commandSet["diagnostic-report"] = commands.NewDiagnosticReport(presenter, api)
	commandSet["disable-director-verifiers"] = commands.NewDisableDirectorVerifiers(presenter, api, stdout)
	commandSet["disable-product-verifiers"] = commands.NewDisableProductVerifiers(presenter, api, stdout)
	commandSet["download-product"] = commands.NewDownloadProduct(os.Environ, stdout, stderr, os.Stderr, api)
	commandSet["errands"] = commands.NewErrands(presenter, api)
	commandSet["expiring-certificates"] = commands.NewExpiringCertificates(api, stdout)
	commandSet["export-installation"] = commands.NewExportInstallation(api, stderr)
	commandSet["generate-certificate"] = commands.NewGenerateCertificate(api, stdout)
	commandSet["generate-certificate-authority"] = commands.NewGenerateCertificateAuthority(api, presenter)
	commandSet["help"] = commands.NewHelp(os.Stdout, globalFlagsUsage, commandSet)
	commandSet["import-installation"] = commands.NewImportInstallation(form, api, global.DecryptionPassphrase, stdout)
	commandSet["installation-log"] = commands.NewInstallationLog(api, stdout)
	commandSet["installations"] = commands.NewInstallations(api, presenter)
	commandSet["interpolate"] = commands.NewInterpolate(os.Environ, stdout, os.Stdin)
	commandSet["pending-changes"] = commands.NewPendingChanges(presenter, api, stderr)
	commandSet["pre-deploy-check"] = commands.NewPreDeployCheck(presenter, api, stdout)
	commandSet["product-metadata"] = commands.NewProductMetadata(stdout)
	commandSet["regenerate-certificates"] = commands.NewRegenerateCertificates(api, stdout)
	commandSet["revert-staged-changes"] = commands.NewRevertStagedChanges(api, stdout)
	commandSet["ssl-certificate"] = commands.NewSSLCertificate(api, presenter)
	commandSet["stage-product"] = commands.NewStageProduct(api, stdout)
	commandSet["staged-config"] = commands.NewStagedConfig(api, stdout)
	commandSet["staged-director-config"] = commands.NewStagedDirectorConfig(api, stdout, stderr)
	commandSet["staged-manifest"] = commands.NewStagedManifest(api, stdout)
	commandSet["staged-products"] = commands.NewStagedProducts(presenter, api)
	commandSet["unstage-product"] = commands.NewUnstageProduct(api, stdout)
	commandSet["upload-product"] = commands.NewUploadProduct(form, metadataExtractor, api, stdout)
	commandSet["upload-stemcell"] = commands.NewUploadStemcell(form, api, stdout)
	commandSet["version"] = commands.NewVersion(version, sout)

	err = commandSet.Execute(command, args)
	if err != nil {
		return err
	}

	return nil
}

func setEnvFileProperties(global *options) error {
	if global.Env == "" {
		return nil
	}

	var opts options
	_, err := os.Open(global.Env)
	if err != nil {
		return fmt.Errorf("env file does not exist: %s", err)
	}

	interpolateOptions := interpolate.Options{
		TemplateFile:  global.Env,
		EnvironFunc:   os.Environ,
		ExpectAllKeys: false,
	}
	if global.VarsEnv != "" {
		interpolateOptions.VarsEnvs = []string{global.VarsEnv}
	}

	contents, err := interpolate.Execute(interpolateOptions)
	if err != nil {
		return err
	}

	err = yaml.UnmarshalStrict(contents, &opts)
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
	if global.ConnectTimeout == 10 && opts.ConnectTimeout != 0 {
		global.ConnectTimeout = opts.ConnectTimeout
	}
	if global.RequestTimeout == 1800 && opts.RequestTimeout != 0 {
		global.RequestTimeout = opts.RequestTimeout
	}
	if !global.SkipSSLValidation {
		global.SkipSSLValidation = opts.SkipSSLValidation
	}
	if global.Target == "" {
		global.Target = opts.Target
	}
	if !global.Trace {
		global.Trace = opts.Trace
	}
	if global.Username == "" {
		global.Username = opts.Username
	}
	if global.DecryptionPassphrase == "" {
		global.DecryptionPassphrase = opts.DecryptionPassphrase
	}
	if global.CACert == "" {
		global.CACert = opts.CACert
	}

	err = checkForVars(global)
	if err != nil {
		return fmt.Errorf("found problem in --env file: %s", err)
	}

	return nil
}

func checkForVars(opts *options) error {
	var errBuffer []string

	interpolateRegex := regexp.MustCompile(`\(\(.*\)\)`)

	if interpolateRegex.MatchString(opts.DecryptionPassphrase) {
		errBuffer = append(errBuffer, "* use OM_DECRYPTION_PASSPHRASE environment variable for the decryption-passphrase value")
	}

	if interpolateRegex.MatchString(opts.ClientID) {
		errBuffer = append(errBuffer, "* use OM_CLIENT_ID environment variable for the client-id value")
	}

	if interpolateRegex.MatchString(opts.ClientSecret) {
		errBuffer = append(errBuffer, "* use OM_CLIENT_SECRET environment variable for the client-secret value")
	}

	if interpolateRegex.MatchString(opts.Password) {
		errBuffer = append(errBuffer, "* use OM_PASSWORD environment variable for the password value")
	}

	if interpolateRegex.MatchString(opts.Target) {
		errBuffer = append(errBuffer, "* use OM_TARGET environment variable for the target value")
	}

	if interpolateRegex.MatchString(opts.Username) {
		errBuffer = append(errBuffer, "* use OM_USERNAME environment variable for the username value")
	}

	if len(errBuffer) > 0 {
		errBuffer = append([]string{"env file contains YAML placeholders. Pleases provide them via interpolation or environment variables."}, errBuffer...)
		errBuffer = append(errBuffer, "Or, to enable interpolation of env.yml with variables from env-vars,")
		errBuffer = append(errBuffer, "set the OM_VARS_ENV env var and put export the needed vars.")

		return fmt.Errorf(strings.Join(errBuffer, "\n"))
	}

	return nil
}
