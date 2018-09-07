package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
		ConfigFile           string `                             long:"config"                short:"c"                    description:"path to yml file containing authentication configuration"`
		Username             string `yaml:"username"              long:"username"              short:"u"  env:"OM_USERNAME" description:"admin username"`
		Password             string `yaml:"password"              long:"password"              short:"p"  env:"OM_PASSWORD" description:"admin password"`
		DecryptionPassphrase string `yaml:"decryption-passphrase" long:"decryption-passphrase" short:"dp"                   description:"passphrase used to encrypt the installation"`
		HTTPProxyURL         string `yaml:"http-proxy-url"        long:"http-proxy-url"                                     description:"proxy for outbound HTTP network traffic"`
		HTTPSProxyURL        string `yaml:"https-proxy-url"       long:"https-proxy-url"                                    description:"proxy for outbound HTTPS network traffic"`
		NoProxy              string `yaml:"no-proxy"              long:"no-proxy"                                           description:"comma-separated list of hosts that do not go through the proxy"`
	}
}

func NewConfigureAuthentication(service configureAuthenticationService, logger logger) ConfigureAuthentication {
	return ConfigureAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureAuthentication) Execute(args []string) error {
	if _, err := jhanda.Parse(&ca.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-authentication flags: %s", err)
	}

	if err := ca.LoadConfigFile(); err != nil {
		return fmt.Errorf("could not parse configure-authentication flags: %s", err)
	}

	if err := ca.ValidateRequiredProperties(); err != nil {
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
		Description:      "This unauthenticated command helps setup the internal userstore authentication mechanism for your Ops Manager.",
		ShortDescription: "configures Ops Manager with an internal userstore and admin user account",
		Flags:            ca.Options,
	}
}

func (ca *ConfigureAuthentication) LoadConfigFile() error {
	if ca.Options.ConfigFile == "" {
		return nil
	}

	configContent, err := ioutil.ReadFile(ca.Options.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %s", ca.Options.ConfigFile, err)
	}

	options := ConfigureAuthentication{}
	if err := yaml.Unmarshal(configContent, &options.Options); err != nil {
		return err
	}

	if ca.Options.Username == "" {
		ca.Options.Username = options.Options.Username
	}
	if ca.Options.Password == "" {
		ca.Options.Password = options.Options.Password
	}
	if ca.Options.DecryptionPassphrase == "" {
		ca.Options.DecryptionPassphrase = options.Options.DecryptionPassphrase
	}
	if ca.Options.HTTPProxyURL == "" {
		ca.Options.HTTPProxyURL = options.Options.HTTPProxyURL
	}
	if ca.Options.HTTPSProxyURL == "" {
		ca.Options.HTTPSProxyURL = options.Options.HTTPSProxyURL
	}
	if ca.Options.NoProxy == "" {
		ca.Options.NoProxy = options.Options.NoProxy
	}

	return nil
}

func (ca ConfigureAuthentication) ValidateRequiredProperties() error {
	if ca.Options.Username == "" {
		return fmt.Errorf("missing required flag \"--username\"")
	}
	if ca.Options.Password == "" {
		return fmt.Errorf("missing required flag \"--password\"")
	}
	if ca.Options.DecryptionPassphrase == "" {
		return fmt.Errorf("missing required flag \"--decryption-passphrase\"")
	}
	return nil
}
