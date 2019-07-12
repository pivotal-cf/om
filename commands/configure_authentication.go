package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

//go:generate counterfeiter -o ./fakes/configure_authentication_service.go --fake-name ConfigureAuthenticationService . configureAuthenticationService
type configureAuthenticationService interface {
	Setup(api.SetupInput) (api.SetupOutput, error)
	EnsureAvailability(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error)
	Info() (api.Info, error)
}

type ConfigureAuthentication struct {
	service configureAuthenticationService
	logger  logger
	Options struct {
		ConfigFile             string   `long:"config"                short:"c"                    description:"path to yml file for configuration (keys must match the following command line flags)"`
		Username               string   `long:"username"              short:"u"  env:"OM_USERNAME" description:"admin username" required:"true"`
		Password               string   `long:"password"              short:"p"  env:"OM_PASSWORD" description:"admin password" required:"true"`
		DecryptionPassphrase   string   `long:"decryption-passphrase" short:"dp"                   description:"passphrase used to encrypt the installation" required:"true"`
		HTTPProxyURL           string   `long:"http-proxy-url"                                     description:"proxy for outbound HTTP network traffic"`
		HTTPSProxyURL          string   `long:"https-proxy-url"                                    description:"proxy for outbound HTTPS network traffic"`
		NoProxy                string   `long:"no-proxy"                                           description:"comma-separated list of hosts that do not go through the proxy"`
		PrecreatedClientSecret string   `long:"precreated-client-secret"                           description:"create a UAA client on the Ops Manager vm. The client_secret will be the value provided to this option"`
		VarsEnv                []string `env:"OM_VARS_ENV" experimental:"true" description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
	}
}

func NewConfigureAuthentication(service configureAuthenticationService, logger logger) ConfigureAuthentication {
	return ConfigureAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureAuthentication) Execute(args []string) error {
	var opsManUaaClientMsg string

	err := loadConfigFile(args, &ca.Options, nil)
	if err != nil {
		return fmt.Errorf("could not parse configure-authentication flags: %s", err)
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

	ca.logger.Printf("configuring internal userstore...")

	input := api.SetupInput{
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
	}

	info, err := ca.service.Info()
	if err != nil {
		return err
	}

	versionAtLeast25, err := info.VersionAtLeast(2, 5)
	if err != nil {
		return err
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
	ca.logger.Printf(opsManUaaClientMsg)

	return nil
}

func (ca ConfigureAuthentication) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command helps setup the internal userstore authentication mechanism for your Ops Manager.",
		ShortDescription: "configures Ops Manager with an internal userstore and admin user account",
		Flags:            ca.Options,
	}
}
