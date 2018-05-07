package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/jhanda"
	"fmt"
	"errors"
)

//go:generate counterfeiter -o ./fakes/configure_authentication_service.go --fake-name ConfigureAuthenticationService . configureAuthenticationService
type configureAuthenticationService interface {
	Setup(api.SetupInput) (api.SetupOutput, error)
	EnsureAvailability(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error)
}

type ConfigureAuthentication struct {
	service configureAuthenticationService
	logger  logger
	Options struct {
		Username             string `long:"username"              short:"u"                  description:"Internal Authentication: admin username"`
		Password             string `long:"password"              short:"p"                  description:"Internal Authentication: admin password"`
		DecryptionPassphrase string `long:"decryption-passphrase" short:"dp" required:"true" description:"passphrase used to encrypt the installation"`
		HTTPProxyURL         string `long:"http-proxy-url"                                   description:"proxy for outbound HTTP network traffic"`
		HTTPSProxyURL        string `long:"https-proxy-url"                                  description:"proxy for outbound HTTPS network traffic"`
		NoProxy              string `long:"no-proxy"                                         description:"comma-separated list of hosts that do not go through the proxy"`
		IDPMetadata          string `long:"saml-idp-metadata"                                description:"SAML Authentication: XML, or URL to XML, for the IDP that Ops Manager should use"`
		BoshIDPMetadata      string `long:"saml-bosh-idp-metadata"                           description:"SAML Authentication: XML, or URL to XML, for the IDP that BOSH should use"`
		RBACAdminGroup       string `long:"saml-rbac-admin-group"                            description:"SAML Authentication: If SAML is specified, please provide the admin group for your SAML"`
		RBACGroupsAttribute  string `long:"saml-rbac-groups-attribute"                       description:"SAML Authentication: If SAML is specified, please provide the groups attribute for your SAML"`
	}
}

func NewConfigureAuthentication(service configureAuthenticationService, logger logger) ConfigureAuthentication {
	return ConfigureAuthentication{
		service: service,
		logger:  logger,
	}
}

type method int

const (
	noMethod method = iota
	internal
	saml
)

func (ca ConfigureAuthentication) validate() (method, error) {
	if ca.Options.Username != "" || ca.Options.Password != "" {
		// expect all saml not set, other wise: Error: should only o=use internal or saml
		// expect all internal setother wise, Error, missing field
		switch {
		case ca.Options.Username == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--username\"")
		case ca.Options.Password == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--password\"")
		case ca.Options.DecryptionPassphrase == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--decryption-passphrase\"")
		}
		return internal, nil
	} else if ca.Options.IDPMetadata != "" ||
		ca.Options.BoshIDPMetadata != "" ||
		ca.Options.RBACAdminGroup != "" ||
		ca.Options.RBACGroupsAttribute != "" {
		// expect all internal not set,  other wise: Error: should only o=use internal or saml
		// expect all saml set, other wise: Error, missing field
		switch {
		case ca.Options.IDPMetadata == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--saml-idp-metadata\"")
		case ca.Options.BoshIDPMetadata == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--saml-bosh-idp-metadata\"")
		case ca.Options.RBACAdminGroup == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--saml-rbac-admin-group\"")
		case ca.Options.RBACGroupsAttribute == "":
			return noMethod, fmt.Errorf("could not parse configure-authentication flags: missing required flag \"--saml-rbac-groups-attribute\"")
		}
		return saml, nil
	}

	return noMethod, fmt.Errorf("No values set!")
}

func (ca ConfigureAuthentication) Execute(args []string) error {
	if _, err := jhanda.Parse(&ca.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-authentication flags: %s", err)
	}

	authType, err := ca.validate()
	if err != nil {
		return err
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


	switch authType {
	case saml:
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
		})
		if err != nil {
			return fmt.Errorf("could not configure authentication: %s", err)
		}
	case internal:
		ca.logger.Printf("configuring internal userstore...")
		// Check for all fields

		_, err = ca.service.Setup(api.SetupInput{
			IdentityProvider:                 "internal",
			AdminUserName:                    ca.Options.Username,
			AdminPassword:                    ca.Options.Password,
			AdminPasswordConfirmation:        ca.Options.Password,
			DecryptionPassphrase:             ca.Options.DecryptionPassphrase,
			DecryptionPassphraseConfirmation: ca.Options.DecryptionPassphrase,
			HTTPProxyURL:                     ca.Options.HTTPProxyURL,
			HTTPSProxyURL:                    ca.Options.HTTPSProxyURL,
			NoProxy:                          ca.Options.NoProxy,
			EULAAccepted:                     "true",
		})
		if err != nil {
			return fmt.Errorf("could not configure authentication: %s", err)
		}
	}

	ca.logger.Printf("waiting for configuration to complete...")
	for ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusComplete {
		ensureAvailabilityOutput, err = ca.service.EnsureAvailability(api.EnsureAvailabilityInput{})
		if err != nil {
			return fmt.Errorf("could not determine final configuration status: %s", err)
		}
	}

	ca.logger.Printf("configuration complete")

	return nil
}

func (ca ConfigureAuthentication) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager.\nThe \"internal\" userstore mechanism is the only currently supported option.",
		ShortDescription: "configures Ops Manager with an internal userstore and admin user account",
		Flags:            ca.Options,
	}
}
