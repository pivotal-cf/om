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
		ConfigFile            string `long:"config"                short:"c"  description:"path to yml file for configuration (keys must match the following command line flags)"`
		DecryptionPassphrase  string `long:"decryption-passphrase" short:"dp" required:"true" description:"passphrase used to encrypt the installation"`
		HTTPProxyURL          string `long:"http-proxy-url"                                   description:"proxy for outbound HTTP network traffic"`
		HTTPSProxyURL         string `long:"https-proxy-url"                                  description:"proxy for outbound HTTPS network traffic"`
		NoProxy               string `long:"no-proxy"                                         description:"comma-separated list of hosts that do not go through the proxy"`
		IDPMetadata           string `long:"saml-idp-metadata"                required:"true" description:"XML, or URL to XML, for the IDP that Ops Manager should use"`
		BoshIDPMetadata       string `long:"saml-bosh-idp-metadata"           required:"true" description:"XML, or URL to XML, for the IDP that BOSH should use"`
		RBACAdminGroup        string `long:"saml-rbac-admin-group"            required:"true" description:"If SAML is specified, please provide the admin group for your SAML"`
		RBACGroupsAttribute   string `long:"saml-rbac-groups-attribute"       required:"true" description:"If SAML is specified, please provide the groups attribute for your SAML"`
		CreateBoshAdminClient bool   `long:"create-bosh-admin-client"                         description:"create a UAA client on the Bosh Director, whose credentials can be passed to the BOSH CLI to execute BOSH commands. Default is false."`
	}
}

func NewConfigureSAMLAuthentication(service configureAuthenticationService, logger logger) ConfigureSAMLAuthentication {
	return ConfigureSAMLAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureSAMLAuthentication) Execute(args []string) error {
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

	if ca.Options.CreateBoshAdminClient == true {
		info, err := ca.service.Info()
		if err != nil {
			return err
		}

		versionAtLeast24, err := info.VersionAtLeast(2, 4)
		if err != nil {
			return err
		}

		if !versionAtLeast24 {
			return fmt.Errorf("create-bosh-client is not supported in OpsMan versions before 2.4")
		}
	}

	ca.logger.Printf("configuring SAML authentication...")

	_, err = ca.service.Setup(api.SetupInput{
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
		CreateBoshAdminClient:            boolStringFromType(ca.Options.CreateBoshAdminClient),
	})
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

	if ca.Options.CreateBoshAdminClient {
		ca.logger.Printf(`
BOSH admin client created.
The new clients secret can be found by going to the OpsMan UI -> director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
Client ID should be 'bosh_admin_client'.
`)
	}

	return nil
}

func (ca ConfigureSAMLAuthentication) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager with SAML.",
		ShortDescription: "configures Ops Manager with SAML authentication",
		Flags:            ca.Options,
	}
}
