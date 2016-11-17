package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type setupService interface {
	Setup(api.SetupInput) (api.SetupOutput, error)
	EnsureAvailability(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error)
}

type ConfigureAuthentication struct {
	service setupService
	logger  logger
	Options struct {
		Username             string `short:"u"  long:"username"              description:"admin username"`
		Password             string `short:"p"  long:"password"              description:"admin password"`
		DecryptionPassphrase string `short:"dp" long:"decryption-passphrase" description:"passphrase used to encrypt the installation"`
	}
}

func NewConfigureAuthentication(service setupService, logger logger) ConfigureAuthentication {
	return ConfigureAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureAuthentication) Usage() Usage {
	return Usage{
		Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager.\nThe \"internal\" userstore mechanism is the only currently supported option.",
		ShortDescription: "configures Ops Manager with an internal userstore and admin user account",
		Flags:            ca.Options,
	}
}

func (ca ConfigureAuthentication) Execute(args []string) error {
	_, err := flags.Parse(&ca.Options, args)
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
	_, err = ca.service.Setup(api.SetupInput{
		IdentityProvider:                 "internal",
		AdminUserName:                    ca.Options.Username,
		AdminPassword:                    ca.Options.Password,
		AdminPasswordConfirmation:        ca.Options.Password,
		DecryptionPassphrase:             ca.Options.DecryptionPassphrase,
		DecryptionPassphraseConfirmation: ca.Options.DecryptionPassphrase,
		EULAAccepted:                     true,
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
