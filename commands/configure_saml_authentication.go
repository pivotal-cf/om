package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureSAMLAuthentication struct {
	service configureAuthenticationService
	logger  logger
	Options struct {
		ConfigFile                string   `long:"config"                short:"c"  description:"path to yml file for configuration (keys must match the following command line flags)"`
		DecryptionPassphrase      string   `long:"decryption-passphrase" short:"dp" required:"true" description:"passphrase used to encrypt the installation"`
		HTTPProxyURL              string   `long:"http-proxy-url"                                   description:"proxy for outbound HTTP network traffic"`
		HTTPSProxyURL             string   `long:"https-proxy-url"                                  description:"proxy for outbound HTTPS network traffic"`
		NoProxy                   string   `long:"no-proxy"                                         description:"comma-separated list of hosts that do not go through the proxy"`
		IDPMetadata               string   `long:"saml-idp-metadata"                required:"true" description:"XML, or URL to XML, for the IDP that Ops Manager should use"`
		BoshIDPMetadata           string   `long:"saml-bosh-idp-metadata"           required:"true" description:"XML, or URL to XML, for the IDP that BOSH should use"`
		RBACAdminGroup            string   `long:"saml-rbac-admin-group"            required:"true" description:"If SAML is specified, please provide the admin group for your SAML"`
		RBACGroupsAttribute       string   `long:"saml-rbac-groups-attribute"       required:"true" description:"If SAML is specified, please provide the groups attribute for your SAML"`
		SkipCreateBoshAdminClient bool     `long:"skip-create-bosh-admin-client"                    description:"create a UAA client on the Bosh Director, whose credentials can be passed to the BOSH CLI to execute BOSH commands. Default is false."`
		PrecreatedClientSecret    string   `long:"precreated-client-secret"                         description:"create a UAA client on the Ops Manager vm, whose secret will be the value provided to this option"`
		VarsEnv                   []string `env:"OM_VARS_ENV" experimental:"true" description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
	}
}

func NewConfigureSAMLAuthentication(service configureAuthenticationService, logger logger) ConfigureSAMLAuthentication {
	return ConfigureSAMLAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureSAMLAuthentication) Execute(args []string) error {
	var (
		boshAdminClientMsg string
		opsManUaaClientMsg string
	)

	err := loadConfigFile(args, &ca.Options, nil)
	if err != nil {
		return fmt.Errorf("could not parse configure-saml-authentication flags: %s", err)
	}

	ensureAvailabilityOutput, err := ca.service.EnsureAvailability(api.EnsureAvailabilityInput{})
	if err != nil {
		return fmt.Errorf("could not determine initial configuration status: %s", err)
	}

	if ensureAvailabilityOutput.Status == api.EnsureAvailabilityStatusUnknown {
		return errors.New("could not determine initial configuration status: received unexpected status")
	}

	if ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusUnstarted {
		ca.logger.Printf("configuration previously completed, skipping configuration")
		return nil
	}

	ca.logger.Printf("configuring SAML authentication...")

	input := api.SetupInput{
		IdentityProvider:                 "saml",
		DecryptionPassphrase:             ca.Options.DecryptionPassphrase,
		DecryptionPassphraseConfirmation: ca.Options.DecryptionPassphrase,
		HTTPProxyURL:                     ca.Options.HTTPProxyURL,
		HTTPSProxyURL:                    ca.Options.HTTPSProxyURL,
		NoProxy:                          ca.Options.NoProxy,
		EULAAccepted:                     "true",
		IDPMetadata:                      ca.Options.IDPMetadata,
		BoshIDPMetadata:                  ca.Options.BoshIDPMetadata,
		RBACAdminGroup:                   ca.Options.RBACAdminGroup,
		RBACGroupsAttribute:              ca.Options.RBACGroupsAttribute,
	}

	info, err := ca.service.Info()
	if err != nil {
		return err
	}

	versionAtLeast24, err := info.VersionAtLeast(2, 4)
	if err != nil {
		return err
	}

	versionAtLeast25, err := info.VersionAtLeast(2, 5)
	if err != nil {
		return err
	}

	if versionAtLeast24 {
		input.CreateBoshAdminClient = !ca.Options.SkipCreateBoshAdminClient
		boshAdminClientMsg = `
BOSH admin client will be created when the director is deployed.
The client secret can then be found in the Ops Manager UI:
director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
`
	} else {
		boshAdminClientMsg = `
Note: BOSH admin client NOT automatically created.
This is only supported in OpsManager 2.4 and up.
`
	}

	if ca.Options.SkipCreateBoshAdminClient {
		boshAdminClientMsg = `
Note: BOSH admin client NOT automatically created.
This was skipped due to the 'skip-create-bosh-admin-client' flag.
`
	}

	if len(ca.Options.PrecreatedClientSecret) > 0 {
		if versionAtLeast25 {
			input.PrecreatedClientSecret = ca.Options.PrecreatedClientSecret
			opsManUaaClientMsg = `
Ops Manager UAA client will be created when authentication system starts.
It will have the username 'precreated-client' and the client secret you provided.
`
		} else {
			return errors.New(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`)
		}
	}

	_, err = ca.service.Setup(input)
	if err != nil {
		return fmt.Errorf("could not configure authentication: %s", err)
	}

	ca.logger.Printf("waiting for configuration to complete...")
	for ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusComplete {
		ensureAvailabilityOutput, err = ca.service.EnsureAvailability(api.EnsureAvailabilityInput{})
		if err != nil {
			return fmt.Errorf("could not determine final configuration status: %s", err)
		}
	}

	ca.logger.Printf("configuration complete")
	ca.logger.Printf(boshAdminClientMsg)
	ca.logger.Printf(opsManUaaClientMsg)

	return nil
}

func (ca ConfigureSAMLAuthentication) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager with SAML.",
		ShortDescription: "configures Ops Manager with SAML authentication",
		Flags:            ca.Options,
	}
}
