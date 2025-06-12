package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/formcontent"
	"github.com/pivotal-cf/om/interpolate"
	"github.com/pivotal-cf/om/network"
	"github.com/pivotal-cf/om/presenters"
	"github.com/pivotal-cf/om/renderers"
	"gopkg.in/yaml.v2"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type options struct {
	CACert               string `yaml:"ca-cert" long:"ca-cert" env:"OM_CA_CERT" description:"OpsManager CA certificate path or value"`
	ClientID             string `yaml:"client-id"             short:"c"  long:"client-id"             env:"OM_CLIENT_ID"                           description:"Client ID for the Ops Manager VM (not required for unauthenticated commands)"`
	ClientSecret         string `yaml:"client-secret"         short:"s"  long:"client-secret"         env:"OM_CLIENT_SECRET"                       description:"Client Secret for the Ops Manager VM (not required for unauthenticated commands)"`
	ConnectTimeout       int    `yaml:"connect-timeout"       short:"o"  long:"connect-timeout"       env:"OM_CONNECT_TIMEOUT"     default:"10"    description:"timeout in seconds to make TCP connections"`
	DecryptionPassphrase string `yaml:"decryption-passphrase" short:"d"  long:"decryption-passphrase" env:"OM_DECRYPTION_PASSPHRASE"               description:"Passphrase to decrypt the installation if the Ops Manager VM has been rebooted (optional for most commands)"`
	Env                  string `                             short:"e"  long:"env"                                                                description:"env file with login credentials"`
	Password             string `yaml:"password"              short:"p"  long:"password"              env:"OM_PASSWORD"                            description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
	RequestTimeout       int    `yaml:"request-timeout"       short:"r"  long:"request-timeout"       env:"OM_REQUEST_TIMEOUT"     default:"1800"  description:"timeout in seconds for HTTP requests to Ops Manager"`
	SkipSSLValidation    bool   `yaml:"skip-ssl-validation"   short:"k"  long:"skip-ssl-validation"   env:"OM_SKIP_SSL_VALIDATION"                 description:"skip ssl certificate validation during http requests"`
	Target               string `yaml:"target"                short:"t"  long:"target"                env:"OM_TARGET"                              description:"location of the Ops Manager VM"`
	UAATarget            string `yaml:"uaa-target"                       long:"uaa-target"            env:"OM_UAA_TARGET"                          description:"optional location of the Ops Manager UAA"`
	Trace                bool   `yaml:"trace"                            long:"trace"                 env:"OM_TRACE"                               description:"prints HTTP requests and response payloads"`
	Username             string `yaml:"username"              short:"u"  long:"username"              env:"OM_USERNAME"                            description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
	VarsEnv              string `                                        long:"vars-env"              env:"OM_VARS_ENV"                            description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
	Version              bool   `                             short:"v"  long:"version"                                                            description:"prints the om release version"`
}

func Main(sout io.Writer, serr io.Writer, version string, applySleepDurationString string, args []string) error {
	applySleepDuration, _ := time.ParseDuration(applySleepDurationString)

	stdout := log.New(sout, "", 0)
	stderr := log.New(serr, "", 0)

	var global options
	parser := flags.NewParser(&global, flags.PassDoubleDash|flags.PassAfterNonOption)
	parser.Name = "om"

	args, _ = parser.ParseArgs(args[1:])

	if global.Version {
		return commands.NewVersion(version, sout).Execute(nil)
	}

	if len(args) > 0 && args[0] == "help" {
		args[0] = "--help"
	}

	err := setEnvFileProperties(&global)
	if err != nil {
		return err
	}

	requestTimeout := time.Duration(global.RequestTimeout) * time.Second
	connectTimeout := time.Duration(global.ConnectTimeout) * time.Second

	var unauthenticatedClient, authedClient, unauthenticatedProgressClient, authedProgressClient httpClient
	unauthenticatedClient, err = network.NewUnauthenticatedClient(global.Target, global.SkipSSLValidation, global.CACert, connectTimeout, requestTimeout)
	if err != nil {
		return err
	}

	authedClient, err = network.NewOAuthClient(global.UAATarget, global.Target, global.Username, global.Password, global.ClientID, global.ClientSecret, global.SkipSSLValidation, global.CACert, connectTimeout, requestTimeout)
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

	command, err := parser.AddCommand(
		"vm-lifecycle",
		"commands to manage the state of the Ops Manager VM",
		"commands to manage the state of the Ops Manager VM. Requires the cli of the desired IAAS to be installed.",
		commands.NewAutomator(os.Stdout, os.Stderr),
	)
	if err != nil {
		return err
	}

	command.Aliases = append(command.Aliases, "nom")

	_, err = parser.AddCommand(
		"activate-certificate-authority",
		"activates a certificate authority on the Ops Manager",
		"This authenticated command activates an existing certificate authority on the Ops Manager",
		commands.NewActivateCertificateAuthority(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"apply-changes",
		"triggers an install on the Ops Manager targeted",
		"This authenticated command kicks off an install of any staged changes on the Ops Manager.",
		commands.NewApplyChanges(api, api, logWriter, stdout, applySleepDuration),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"assign-multi-stemcell",
		"assigns multiple uploaded stemcells to a product in the targeted Ops Manager 2.6+",
		"This command will assign multiple already uploaded stemcells to a specific product in Ops Manager 2.6+.\n"+
			"It is recommended to use \"upload-stemcell --floating=false\" before using this command.",
		commands.NewAssignMultiStemcell(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"assign-stemcell",
		"assigns an uploaded stemcell to a product in the targeted Ops Manager",
		"This command will assign an already uploaded stemcell to a specific product in Ops Manager.\n"+
			"It is recommended to use \"upload-stemcell --floating=false\" before using this command.",
		commands.NewAssignStemcell(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"available-products",
		"**DEPRECATED** lists available products. Use 'products --available' instead.",
		"**DEPRECATED** This authenticated command lists all available products. Use 'products --available' instead.",
		commands.NewAvailableProducts(api, presenter, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"bosh-diff",
		"displays BOSH manifest diff for the director and products",
		"This command displays the bosh manifest diff for the director and products (Note: secret values are replaced with double-paren variable names)",
		commands.NewBoshDiff(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"bosh-env",
		"prints environment variables for BOSH and Credhub",
		"This prints environment variables to target the BOSH director and Credhub. You can invoke it directly to see its output, or use it directly with an evaluate-type command:\nOn posix system: eval \"$(om bosh-env)\"\nOn powershell: iex $(om bosh-env | Out-String)",
		commands.NewBoshEnvironment(api, stdout, global.Target, envRendererFactory),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"certificate-authorities",
		"lists certificates managed by Ops Manager",
		"lists certificates managed by Ops Manager",
		commands.NewCertificateAuthorities(api, presenter),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"certificate-authority",
		"prints requested certificate authority",
		"prints requested certificate authority",
		commands.NewCertificateAuthority(api, presenter, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"config-template",
		"generates a config template from a Pivnet product",
		"this command generates a product configuration template from a .pivotal file or Pivnet",
		commands.NewConfigTemplate(commands.DefaultConfigTemplateProvider()),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-authentication",
		"configures Ops Manager with an internal userstore and admin user account",
		"This unauthenticated command helps setup the internal userstore authentication mechanism for your Ops Manager.",
		commands.NewConfigureAuthentication(os.Environ, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-director",
		"configures the director",
		"This authenticated command configures the director.",
		commands.NewConfigureDirector(os.Environ, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-ldap-authentication",
		"configures Ops Manager with LDAP authentication",
		"This unauthenticated command helps setup the authentication mechanism for your Ops Manager with LDAP.",
		commands.NewConfigureLDAPAuthentication(os.Environ, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-opsman",
		"configures values present on the Ops Manager settings page",
		"This authenticated command configures settings available on the \"Settings\" page in the Ops Manager UI. For an example config, reference the docs directory for this command.",
		commands.NewConfigureOpsman(os.Environ, api, stderr),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-product",
		"configures a staged product",
		"This authenticated command configures a staged product",
		commands.NewConfigureProduct(os.Environ, api, global.Target, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"configure-saml-authentication",
		"configures Ops Manager with SAML authentication",
		"This unauthenticated command helps setup the authentication mechanism for your Ops Manager with SAML.",
		commands.NewConfigureSAMLAuthentication(os.Environ, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"create-certificate-authority",
		"creates a certificate authority on the Ops Manager",
		"This authenticated command creates a certificate authority on the Ops Manager with the given cert and key",
		commands.NewCreateCertificateAuthority(api, presenter),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"create-vm-extension",
		"creates/updates a VM extension",
		"This creates/updates a VM extension",
		commands.NewCreateVMExtension(os.Environ, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"credential-references",
		"list credential references for a deployed product",
		"This authenticated command lists credential references for deployed products.",
		commands.NewCredentialReferences(api, presenter, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"credentials",
		"fetch credentials for a deployed product",
		"This authenticated command fetches credentials for deployed products.",
		commands.NewCredentials(api, presenter, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"curl",
		"issues an authenticated API request",
		"This command issues an authenticated API request as defined in the arguments",
		commands.NewCurl(api, stdout, stderr),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"delete-certificate-authority",
		"deletes a certificate authority on the Ops Manager",
		"This authenticated command deletes an existing certificate authority on the Ops Manager",
		commands.NewDeleteCertificateAuthority(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"delete-installation",
		"deletes all the products on the Ops Manager targeted",
		"This authenticated command deletes all the products installed on the targeted Ops Manager.",
		commands.NewDeleteInstallation(api, logWriter, stdout, os.Stdin, applySleepDuration),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"delete-product",
		"deletes an unused product from the Ops Manager",
		"This command deletes the specified unused product from the targeted Ops Manager",
		commands.NewDeleteProduct(api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"delete-ssl-certificate",
		"deletes certificate applied to Ops Manager",
		"This authenticated command deletes a custom certificate applied to Ops Manager and reverts to the auto-generated cert",
		commands.NewDeleteSSLCertificate(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"delete-unused-products",
		"deletes unused products on the Ops Manager targeted",
		"This command deletes unused products in the targeted Ops Manager",
		commands.NewDeleteUnusedProducts(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"deployed-manifest",
		"prints the deployed manifest for a product",
		"This authenticated command prints the deployed manifest for a product",
		commands.NewDeployedManifest(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"deployed-products",
		"**DEPRECATED** lists deployed products. Use 'products --deployed' instead.",
		"**DEPRECATED** This authenticated command lists all deployed products. Use 'products --deployed' instead.",
		commands.NewDeployedProducts(presenter, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"diagnostic-report",
		"reports current state of your Ops Manager",
		"retrieve a diagnostic report with general information about the state of your Ops Manager.",
		commands.NewDiagnosticReport(presenter, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"disable-director-verifiers",
		"disables director verifiers",
		"This authenticated command disables director verifiers",
		commands.NewDisableDirectorVerifiers(presenter, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"disable-product-verifiers",
		"disables product verifiers",
		"This authenticated command disables product verifiers",
		commands.NewDisableProductVerifiers(presenter, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"download-product",
		"downloads a specified product file from Pivotal Network",
		"This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
		commands.NewDownloadProduct(os.Environ, stdout, stderr, os.Stderr, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"errands",
		"list errands for a product",
		"This authenticated command lists all errands for a product.",
		commands.NewErrands(presenter, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"expiring-certificates",
		"lists expiring certificates from the Ops Manager targeted",
		"returns a list of expiring certificates from an existing Ops Manager",
		commands.NewExpiringCertificates(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"expiring-licenses",
		"lists expiring licenses from the Ops Manager targeted",
		"returns a list of expiring licenses from an existing Ops Manager",
		commands.NewExpiringLicenses(presenter, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"export-installation",
		"exports the installation of the target Ops Manager",
		"This command will export the current installation of the target Ops Manager.",
		commands.NewExportInstallation(api, stderr),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"generate-certificate",
		"generates a new certificate signed by Ops Manager's root CA",
		"This authenticated command generates a new RSA public/private certificate signed by Ops Manager’s root CA certificate",
		commands.NewGenerateCertificate(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"generate-certificate-authority",
		"generates a certificate authority on the Opsman",
		"This authenticated command generates a certificate authority on the Ops Manager",
		commands.NewGenerateCertificateAuthority(api, presenter),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"import-installation",
		"imports a given installation to the Ops Manager targeted",
		"This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
		commands.NewImportInstallation(form, api, global.DecryptionPassphrase, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"installation-log",
		"output installation logs",
		"This authenticated command retrieves the logs for a given installation.",
		commands.NewInstallationLog(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"installations",
		"list recent installation events",
		"This authenticated command lists all recent installation events.",
		commands.NewInstallations(api, presenter),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"interpolate",
		"interpolates variables into a manifest",
		"interpolates variables into a manifest",
		commands.NewInterpolate(os.Environ, stdout, os.Stdin),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"pending-changes",
		"checks for pending changes",
		"This authenticated command lists all products and will display whether they are unchanged (no pending changes) or changed (has pending changes).",
		commands.NewPendingChanges(presenter, api, stderr),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"pre-deploy-check",
		"checks completeness and validity of product configuration",
		"This authenticated checks completeness and validity of product configuration. This includes whether stemcells are assigned, missing configuration, and failed validators for a product.",
		commands.NewPreDeployCheck(presenter, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"product-metadata",
		"prints product metadata",
		"This command prints metadata about the given product from a .pivotal file or Pivnet",
		commands.NewProductMetadata(commands.DefaultProductMetadataProvider(), stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"products",
		"lists product staged, available, and deployed versions",
		"This authenticated command lists all products. Staged, available, and deployed are listed by default.",
		commands.NewProducts(presenter, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"regenerate-certificates",
		"deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes",
		"This authenticated command deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes",
		commands.NewRegenerateCertificates(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"revert-staged-changes",
		"This command reverts the staged changes already on an Ops Manager.",
		"This command reverts the staged changes already on an Ops Manager. Useful for ensuring that unintended changes are not applied.",
		commands.NewRevertStagedChanges(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"ssl-certificate",
		"gets certificate applied to Ops Manager",
		"This authenticated command gets certificate applied to Ops Manager",
		commands.NewSSLCertificate(api, presenter),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"stage-product",
		"stages a given product in the Ops Manager targeted",
		"This command attempts to stage a product in the Ops Manager",
		commands.NewStageProduct(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"staged-config",
		"generates a config from a staged product",
		"This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
		commands.NewStagedConfig(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"staged-director-config",
		"generates a config from a staged director",
		"This command generates a config from a staged director that can be passed in to om configure-director",
		commands.NewStagedDirectorConfig(api, stdout, stderr),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"staged-manifest",
		"prints the staged manifest for a product",
		"This authenticated command prints the staged manifest for a product",
		commands.NewStagedManifest(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"staged-products",
		"**DEPRECATED** lists staged products. Use 'products --staged' instead.",
		"**DEPRECATED** This authenticated command lists all staged products. Use 'products --staged' instead.",
		commands.NewStagedProducts(presenter, api),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"unstage-product",
		"unstages a given product from the Ops Manager targeted",
		"This command attempts to unstage a product from the Ops Manager",
		commands.NewUnstageProduct(api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"upload-product",
		"uploads a given product to the Ops Manager targeted",
		"This command attempts to upload a product to the Ops Manager",
		commands.NewUploadProduct(form, metadataExtractor, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"upload-stemcell",
		"uploads a given stemcell to the Ops Manager targeted",
		"This command will upload a stemcell to the target Ops Manager. Unless the force flag is used, if the stemcell already exists that upload will be skipped",
		commands.NewUploadStemcell(form, api, stdout),
	)
	if err != nil {
		return err
	}
	_, err = parser.AddCommand(
		"version",
		"prints the om release version",
		"This command prints the om release version number.",
		commands.NewVersion(version, sout),
	)
	if err != nil {
		return err
	}

	args, err = loadConfigFile(args, os.Environ)
	if err != nil {
		return err
	}

	// Strict flag validation block
	err = validateCommandFlags(parser, args)
	if err != nil {
		return err
	}

	parser.Options |= flags.HelpFlag
	if _, err = parser.ParseArgs(args); err != nil {
		if e, ok := err.(*flags.Error); ok {
			switch e.Type {
			case flags.ErrHelp, flags.ErrCommandRequired:
				parser.WriteHelp(os.Stdout)
				return nil
			}
		}
	}

	return err
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

		return errors.New(strings.Join(errBuffer, "\n"))
	}

	return nil
}

// Helper to recursively collect all options from a group and its subgroups
func collectAllOptions(group *flags.Group) []*flags.Option {
	var allOptions []*flags.Option
	allOptions = append(allOptions, group.Options()...)
	for _, subgroup := range group.Groups() {
		allOptions = append(allOptions, collectAllOptions(subgroup)...)
	}
	return allOptions
}

// validateCommandFlags checks if the provided command flags are valid for the given command.
func validateCommandFlags(parser *flags.Parser, args []string) error {
	if len(args) == 0 {
		return nil
	}

	var selectedCmd *flags.Command
	cmds := parser.Commands()
	// currentArgPosition tracks our position in the args array as we traverse the command hierarchy
	currentArgPosition := 0

	// Find the top-level command
	for _, cmd := range cmds {
		if cmd.Name == args[currentArgPosition] || contains(cmd.Aliases, args[currentArgPosition]) {
			selectedCmd = cmd
			break
		}
	}
	if selectedCmd == nil {
		return nil // unknown command, let parser handle it
	}
	currentArgPosition++

	// Walk down subcommands as long as the next arg matches a subcommand
	// For example: "om vm-lifecycle export-opsman-config" would traverse:
	// 1. "vm-lifecycle" (top-level command)
	// 2. "export-opsman-config" (subcommand)
	for currentArgPosition < len(args) {
		found := false
		for _, sub := range selectedCmd.Commands() {
			if sub.Name == args[currentArgPosition] || contains(sub.Aliases, args[currentArgPosition]) {
				selectedCmd = sub
				currentArgPosition++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Now selectedCmd is the deepest subcommand
	// Check remaining args for unknown flags
	invalidFlags := findUnknownFlags(selectedCmd, args[currentArgPosition:])

	if len(invalidFlags) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown flag(s) %q for command '%s'\n", invalidFlags, selectedCmd.Name)
		fmt.Fprintf(os.Stderr, "See 'om %s --help' for available options.\n", selectedCmd.Name)
		os.Exit(1)
	}
	return nil
}

// findUnknownFlags checks for unknown flags in the provided args for the given command.
func findUnknownFlags(selectedCmd *flags.Command, args []string) []string {
	validFlags := make(map[string]bool)
	addFlag := func(name string, takesValue bool) {
		validFlags[name] = takesValue
	}
	cmd := selectedCmd
	for cmd.Active != nil {
		cmd = cmd.Active
	}
	for _, opt := range collectAllOptions(cmd.Group) {
		val := opt.Value()
		_, isBool := val.(*bool)
		_, isBoolSlice := val.(*[]bool)
		takesValue := !(isBool || isBoolSlice)
		if ln := opt.LongNameWithNamespace(); ln != "" {
			addFlag("--"+ln, takesValue)
		}
		if opt.ShortName != 0 {
			addFlag("-"+string(opt.ShortName), takesValue)
		}
	}
	addFlag("--help", false)
	addFlag("-h", false)

	var invalidFlags []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			// Not a flag, and not a value for a previous flag (since we only check flags)
			// Stop processing further, as all remaining args are positional
			break
		}

		// Split flag and value if --flag=value
		flagName, hasEquals := arg, false
		if eqIdx := strings.Index(arg, "="); eqIdx != -1 {
			flagName = arg[:eqIdx]
			hasEquals = true
			// Example: arg = "--product=foo.pivotal" -> flagName = "--product", value = "foo.pivotal"
		}

		takesValue, isValid := validFlags[flagName]
		if !isValid {
			// Unknown flag
			// Example: arg = "--notaflag" (not defined in command options)
			invalidFlags = append(invalidFlags, flagName)
			i++
			continue
		}

		if takesValue {
			if hasEquals {
				// --flag=value, value is in this arg
				// Example: arg = "--product=foo.pivotal"
				i++
			} else if i+1 < len(args) {
				// --flag value, value is next arg (even if it looks like a flag)
				// Example: args = ["--product", "--notaflag"]
				// "--notaflag" is treated as the value for --product, not as a flag
				i += 2
			} else {
				// --flag with missing value.
				// No need to handle this here as this will handled appropriately by the parser.
				// Example: args = ["--product"] (no value provided)
				i++
			}
		} else {
			// Boolean flag, no value expected
			// Example: arg = "--help"
			i++
		}
	}
	return invalidFlags
}

// contains checks if a string is present in a list of strings.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
