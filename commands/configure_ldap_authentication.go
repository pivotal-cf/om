package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureLDAPAuthentication struct {
	service configureAuthenticationService
	logger  logger
	Options struct {
		ConfigFile           string `long:"config"                short:"c"  description:"path to yml file for configuration (keys must match the following command line flags)"`
		DecryptionPassphrase string `long:"decryption-passphrase" short:"dp" required:"true" description:"passphrase used to encrypt the installation"`
		EmailAttribute       string `long:"email-attribute" required:"true"`
		GroupSearchBase      string `long:"group-search-base" required:"true"`
		GroupSearchFilter    string `long:"group-search-filter" required:"true"`
		LDAPPassword         string `long:"ldap-password" required:"true"`
		LDAPRBACAdminGroup   string `long:"ldap-rbac-admin-group-name" required:"true"`
		LDAPReferral         string `long:"ldap-referrals" required:"true"`
		LDAPUsername         string `long:"ldap-username" required:"true"`
		ServerURL            string `long:"server-url" required:"true"`
		UserSearchBase       string `long:"user-search-base" required:"true"`
		UserSearchFilter     string `long:"user-search-filter" required:"true"`
	}
}

func NewConfigureLDAPAuthentication(service configureAuthenticationService, logger logger) ConfigureLDAPAuthentication {
	return ConfigureLDAPAuthentication{
		service: service,
		logger:  logger,
	}
}

func (ca ConfigureLDAPAuthentication) Execute(args []string) error {
	err := loadConfigFile(args, &ca.Options, nil)
	if err != nil {
		return fmt.Errorf("could not parse configure-ldap-authentication flags: %s", err)
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

	ca.logger.Printf("configuring LDAP authentication...")

	_, err = ca.service.Setup(api.SetupInput{
		IdentityProvider:                 "ldap",
		DecryptionPassphrase:             ca.Options.DecryptionPassphrase,
		DecryptionPassphraseConfirmation: ca.Options.DecryptionPassphrase,
		EULAAccepted:                     "true",
		LDAPSettings: &api.LDAPSettings{
			EmailAttribute:     ca.Options.EmailAttribute,
			GroupSearchBase:    ca.Options.GroupSearchBase,
			GroupSearchFilter:  ca.Options.GroupSearchFilter,
			LDAPPassword:       ca.Options.LDAPPassword,
			LDAPRBACAdminGroup: ca.Options.LDAPRBACAdminGroup,
			LDAPReferral:       ca.Options.LDAPReferral,
			LDAPUsername:       ca.Options.LDAPUsername,
			ServerURL:          ca.Options.ServerURL,
			UserSearchBase:     ca.Options.UserSearchBase,
			UserSearchFilter:   ca.Options.UserSearchFilter,
		},
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

func (ca ConfigureLDAPAuthentication) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager with LDAP.",
		ShortDescription: "configures Ops Manager with LDAP authentication",
		Flags:            ca.Options,
	}
}
