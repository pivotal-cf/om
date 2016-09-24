package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type setupService interface {
	Setup(api.SetupInput) (api.SetupOutput, error)
}

type ConfigureAuthentication struct {
	service setupService
	Options struct {
		Username             string `short:"u"  long:"username"              description:"admin username"`
		Password             string `short:"p"  long:"password"              description:"admin password"`
		DecryptionPassphrase string `short:"dp" long:"decryption-passphrase" description:"passphrase used to encrypt the installation"`
	}
}

func NewConfigureAuthentication(service setupService) ConfigureAuthentication {
	return ConfigureAuthentication{
		service: service,
	}
}

func (ca ConfigureAuthentication) Help() string {
	return "configures OpsManager with an internal userstore and admin user account"
}

func (ca ConfigureAuthentication) Execute(args []string) error {
	_, err := flags.Parse(&ca.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse configure-authentication flags: %s", err)
	}

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

	return nil
}
