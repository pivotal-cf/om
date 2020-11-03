package commands_test

import (
	"errors"
	"github.com/onsi/gomega/gbytes"
	"log"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureLDAPAuthentication.Execute", func() {
	var (
		service             *fakes.ConfigureAuthenticationService
		stdout              *gbytes.Buffer
		logger              *log.Logger
		command             *commands.ConfigureLDAPAuthentication
		commandLineArgs     []string
		expectedPayload     api.SetupInput
		eaExpectedCallCount int
	)

	BeforeEach(func() {
		service = &fakes.ConfigureAuthenticationService{}
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)

		eaOutputs := []api.EnsureAvailabilityOutput{
			{Status: api.EnsureAvailabilityStatusUnstarted},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusComplete},
		}

		eaExpectedCallCount = len(eaOutputs)

		service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
			return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
		}

		command = commands.NewConfigureLDAPAuthentication(nil, service, logger)

		service.InfoReturns(api.Info{
			Version: "2.5-build.1",
		}, nil)

		commandLineArgs = []string{
			"--decryption-passphrase", "some-passphrase",
			"--email-attribute", "mail",
			"--server-url", "ldap://YOUR-LDAP-SERVER",
			"--ldap-username", "cn=admin,dc=opsmanager,dc=com",
			"--ldap-password", "password",
			"--user-search-base", "ou=users,dc=opsmanager,dc=com",
			"--user-search-filter", "cn={0}",
			"--group-search-base", "ou=groups,dc=opsmanager,dc=com",
			"--group-search-filter", "member={0}",
			"--ldap-rbac-admin-group-name", "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
			"--ldap-referrals", "follow",
		}

		expectedPayload = api.SetupInput{
			IdentityProvider:                 "ldap",
			DecryptionPassphrase:             "some-passphrase",
			DecryptionPassphraseConfirmation: "some-passphrase",
			EULAAccepted:                     "true",
			CreateBoshAdminClient:            true,
			LDAPSettings: &api.LDAPSettings{
				EmailAttribute:     "mail",
				GroupSearchBase:    "ou=groups,dc=opsmanager,dc=com",
				GroupSearchFilter:  "member={0}",
				LDAPPassword:       "password",
				LDAPRBACAdminGroup: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
				LDAPReferral:       "follow",
				LDAPUsername:       "cn=admin,dc=opsmanager,dc=com",
				ServerURL:          "ldap://YOUR-LDAP-SERVER",
				UserSearchBase:     "ou=users,dc=opsmanager,dc=com",
				UserSearchFilter:   "cn={0}",
			},
		}
	})

	It("configures LDAP authentication", func() {
		commandLineArgs = append(commandLineArgs, "--precreated-client-secret", "test-client-secret")
		expectedPayload.PrecreatedClientSecret = "test-client-secret"

		err := executeCommand(command,commandLineArgs)
		Expect(err).ToNot(HaveOccurred())

		Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

		Expect(service.EnsureAvailabilityCallCount()).To(Equal(eaExpectedCallCount))

		Expect(stdout).To(gbytes.Say("configuring LDAP authentication..."))
		Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(stdout).To(gbytes.Say("configuration complete"))
		Expect(stdout).To(gbytes.Say(`
BOSH admin client will be created when the director is deployed.
The client secret can then be found in the Ops Manager UI:
director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
`))
		Expect(stdout).To(gbytes.Say(`
Ops Manager UAA client will be created when authentication system starts.
It will have the username 'precreated-client' and the client secret you provided.
`))
	})

	When("OpsMan is < 2.4", func() {
		BeforeEach(func() {
			service.InfoReturns(api.Info{
				Version: "2.3-build.1",
			}, nil)

			expectedPayload.CreateBoshAdminClient = false
			expectedPayload.PrecreatedClientSecret = ""
		})

		It("configure LDAP with bosh admin client warning", func() {
			err := executeCommand(command,commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

			Expect(stdout).To(gbytes.Say("configuring LDAP authentication..."))
			Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
			Expect(stdout).To(gbytes.Say("configuration complete"))
			Expect(stdout).To(gbytes.Say(`
Note: BOSH admin client NOT automatically created.
This is only supported in OpsManager 2.4 and up.
`))
		})
	})

	When("OpsMan is < 2.5", func() {
		BeforeEach(func() {
			service.InfoReturns(api.Info{
				Version: "2.4-build.1",
			}, nil)

			commandLineArgs = append(commandLineArgs, "--precreated-client-secret", "test-client-secret")
		})

		It("errors out if you try to provide a client secret", func() {
			err := executeCommand(command,commandLineArgs)
			Expect(err).To(MatchError(ContainSubstring(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`)))
		})
	})

	When("the skip-create-bosh-admin-client flag is set", func() {
		BeforeEach(func() {
			commandLineArgs = append(commandLineArgs, "--skip-create-bosh-admin-client")
			expectedPayload.CreateBoshAdminClient = false
		})

		It("configures LDAP auth and notifies the user that it skipped client creation", func() {
			err := executeCommand(command,commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

			Expect(stdout).To(gbytes.Say("configuring LDAP authentication..."))
			Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
			Expect(stdout).To(gbytes.Say("configuration complete"))
			Expect(stdout).To(gbytes.Say(`
Note: BOSH admin client NOT automatically created.
This was skipped due to the 'skip-create-bosh-admin-client' flag.
`))
		})

		Context("and OpsMan is < 2.4", func() {
			BeforeEach(func() {
				service.InfoReturns(api.Info{
					Version: "2.3-build.1",
				}, nil)
				commandLineArgs = append(commandLineArgs, "--skip-create-bosh-admin-client")
				expectedPayload.CreateBoshAdminClient = false
				expectedPayload.PrecreatedClientSecret = ""
			})

			It("configures LDAP and notifies the user that it skipped client creation", func() {
				err := executeCommand(command,commandLineArgs)
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(stdout).To(gbytes.Say("configuring LDAP authentication..."))
				Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
				Expect(stdout).To(gbytes.Say("configuration complete"))
				Expect(stdout).To(gbytes.Say(`
Note: BOSH admin client NOT automatically created.
This was skipped due to the 'skip-create-bosh-admin-client' flag.
`))
			})
		})
	})

	When("the authentication setup has already been configured", func() {
		BeforeEach(func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusComplete,
			}, nil)
		})

		It("returns without configuring the authentication system", func() {
			err := executeCommand(command,commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
			Expect(service.SetupCallCount()).To(Equal(0))

			Expect(stdout).To(gbytes.Say("configuration previously completed, skipping configuration"))
		})
	})

	When("a complete config file is provided", func() {
		var configFile *os.File

		BeforeEach(func() {
			var err error
			configContent := `
decryption-passphrase: "some-passphrase"
server-url: "ldap://YOUR-LDAP-SERVER"
ldap-username: "cn=admin,dc=opsmanager,dc=com"
ldap-password: "password"
user-search-base: "ou=users,dc=opsmanager,dc=com"
user-search-filter: "cn={0}"
group-search-base: "ou=groups,dc=opsmanager,dc=com"
group-search-filter: "member={0}"
ldap-rbac-admin-group-name: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com"
email-attribute: "mail"
ldap-referrals: "follow"
precreated-client-secret: test-client-secret
`
			configFile, err = ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			defer configFile.Close()

			_, err = configFile.WriteString(configContent)
			Expect(err).ToNot(HaveOccurred())
			expectedPayload.PrecreatedClientSecret = "test-client-secret"
		})

		It("reads configuration from config file", func() {
			err := executeCommand(command,[]string{
				"--config", configFile.Name(),
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

			Expect(stdout).To(gbytes.Say("configuring LDAP authentication..."))
			Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
			Expect(stdout).To(gbytes.Say("configuration complete"))
		})

		It("is overridden by commandline flags", func() {
			err := executeCommand(command,[]string{
				"--config", configFile.Name(),
				"--server-url", "ldap://example.com",
			})
			Expect(err).ToNot(HaveOccurred())

			expectedPayload.LDAPSettings.ServerURL = "ldap://example.com"

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

		})
	})

	When("the config file contains variables", func() {
		var configFile string

		BeforeEach(func() {
			service.EnsureAvailabilityReturnsOnCall(0, api.EnsureAvailabilityOutput{Status: api.EnsureAvailabilityStatusUnstarted}, nil)
			service.EnsureAvailabilityReturnsOnCall(1, api.EnsureAvailabilityOutput{Status: api.EnsureAvailabilityStatusComplete}, nil)

			configContent := `
decryption-passphrase: ((passphrase))
email-attribute: mail
server-url: ldap://YOUR-LDAP-SERVER
ldap-username: cn=admin,dc=opsmanager,dc=com
ldap-password: ((password))
user-search-base: ou=users,dc=opsmanager,dc=com
user-search-filter: cn={0}
group-search-base: ou=groups,dc=opsmanager,dc=com
group-search-filter: member={0}
ldap-rbac-admin-group-name: cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com
ldap-referrals: follow
`

			configFile = writeTestConfigFile(configContent)
		})

		Context("variables are not provided", func() {
			It("returns an error", func() {
				command := commands.NewConfigureLDAPAuthentication(nil, service, logger)
				err := executeCommand(command,[]string{
					"--config", configFile,
				})
				Expect(err).To(MatchError(ContainSubstring("Expected to find variables")))
			})
		})

		Context("passed in a file (--vars-file)", func() {
			var varsFile string

			BeforeEach(func() {
				varsContent := `
passphrase: a-vars-file-passphrase
password: a-vars-file-password
`

				file, err := ioutil.TempFile("", "vars-*.yml")
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(file.Name(), []byte(varsContent), 0777)
				Expect(err).ToNot(HaveOccurred())

				varsFile = file.Name()
			})

			It("uses values from the vars file", func() {
				command := commands.NewConfigureLDAPAuthentication(nil, service, logger)
				err := executeCommand(command,[]string{
					"--config", configFile,
					"--vars-file", varsFile,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "ldap",
					DecryptionPassphrase:             "a-vars-file-passphrase",
					DecryptionPassphraseConfirmation: "a-vars-file-passphrase",
					EULAAccepted:                     "true",
					CreateBoshAdminClient:            true,
					LDAPSettings: &api.LDAPSettings{
						EmailAttribute:     "mail",
						GroupSearchBase:    "ou=groups,dc=opsmanager,dc=com",
						GroupSearchFilter:  "member={0}",
						LDAPPassword:       "a-vars-file-password",
						LDAPRBACAdminGroup: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
						LDAPReferral:       "follow",
						LDAPUsername:       "cn=admin,dc=opsmanager,dc=com",
						ServerURL:          "ldap://YOUR-LDAP-SERVER",
						UserSearchBase:     "ou=users,dc=opsmanager,dc=com",
						UserSearchFilter:   "cn={0}",
					},
				}))
			})
		})

		Context("passed in a var (--var)", func() {
			It("uses values from the command line", func() {
				command := commands.NewConfigureLDAPAuthentication(nil, service, logger)
				err := executeCommand(command,[]string{
					"--config", configFile,
					"--var", "password=a-command-line-password",
					"--var", "passphrase=a-command-line-passphrase",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "ldap",
					DecryptionPassphrase:             "a-command-line-passphrase",
					DecryptionPassphraseConfirmation: "a-command-line-passphrase",
					EULAAccepted:                     "true",
					CreateBoshAdminClient:            true,
					LDAPSettings: &api.LDAPSettings{
						EmailAttribute:     "mail",
						GroupSearchBase:    "ou=groups,dc=opsmanager,dc=com",
						GroupSearchFilter:  "member={0}",
						LDAPPassword:       "a-command-line-password",
						LDAPRBACAdminGroup: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
						LDAPReferral:       "follow",
						LDAPUsername:       "cn=admin,dc=opsmanager,dc=com",
						ServerURL:          "ldap://YOUR-LDAP-SERVER",
						UserSearchBase:     "ou=users,dc=opsmanager,dc=com",
						UserSearchFilter:   "cn={0}",
					},
				}))
			})
		})

		Context("passed as environment variables (--vars-env)", func() {
			It("interpolates variables into the configuration", func() {
				command := commands.NewConfigureLDAPAuthentication(
					func() []string {
						return []string{"OM_VAR_password=an-env-var-password", "OM_VAR_passphrase=an-env-var-passphrase"}
					},
					service,
					logger,
				)

				err := executeCommand(command,[]string{
					"--config", configFile,
					"--vars-env", "OM_VAR",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "ldap",
					DecryptionPassphrase:             "an-env-var-passphrase",
					DecryptionPassphraseConfirmation: "an-env-var-passphrase",
					EULAAccepted:                     "true",
					CreateBoshAdminClient:            true,
					LDAPSettings: &api.LDAPSettings{
						EmailAttribute:     "mail",
						GroupSearchBase:    "ou=groups,dc=opsmanager,dc=com",
						GroupSearchFilter:  "member={0}",
						LDAPPassword:       "an-env-var-password",
						LDAPRBACAdminGroup: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
						LDAPReferral:       "follow",
						LDAPUsername:       "cn=admin,dc=opsmanager,dc=com",
						ServerURL:          "ldap://YOUR-LDAP-SERVER",
						UserSearchBase:     "ou=users,dc=opsmanager,dc=com",
						UserSearchFilter:   "cn={0}",
					},
				}))
			})

			It("supports the experimental feature of OM_VARS_ENV", func() {
				err := os.Setenv("OM_VARS_ENV", "OM_VAR")
				Expect(err).ToNot(HaveOccurred())
				defer os.Unsetenv("OM_VARS_ENV")

				command := commands.NewConfigureLDAPAuthentication(
					func() []string {
						return []string{"OM_VAR_password=an-env-var-password", "OM_VAR_passphrase=an-env-var-passphrase"}
					},
					service,
					logger,
				)

				err = executeCommand(command,[]string{
					"--config", configFile,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "ldap",
					DecryptionPassphrase:             "an-env-var-passphrase",
					DecryptionPassphraseConfirmation: "an-env-var-passphrase",
					EULAAccepted:                     "true",
					CreateBoshAdminClient:            true,
					LDAPSettings: &api.LDAPSettings{
						EmailAttribute:     "mail",
						GroupSearchBase:    "ou=groups,dc=opsmanager,dc=com",
						GroupSearchFilter:  "member={0}",
						LDAPPassword:       "an-env-var-password",
						LDAPRBACAdminGroup: "cn=opsmgradmins,ou=groups,dc=opsmanager,dc=com",
						LDAPReferral:       "follow",
						LDAPUsername:       "cn=admin,dc=opsmanager,dc=com",
						ServerURL:          "ldap://YOUR-LDAP-SERVER",
						UserSearchBase:     "ou=users,dc=opsmanager,dc=com",
						UserSearchFilter:   "cn={0}",
					},
				}))
			})
		})
	})

	Context("failure cases", func() {
		When("config file cannot be opened", func() {
			It("returns an error", func() {
				err := executeCommand(command,[]string{"--config", "something"})
				Expect(err).To(MatchError("could not parse configure-ldap-authentication flags: could not load the config file: could not read file (something): open something: no such file or directory"))
			})
		})

		When("the initial configuration status cannot be determined", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

				err := executeCommand(command,commandLineArgs)
				Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
			})
		})

		When("the initial configuration status is unknown", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnknown,
				}, nil)

				err := executeCommand(command,commandLineArgs)
				Expect(err).To(MatchError("could not determine initial configuration status: received unexpected status"))
			})
		})

		When("the setup service encounters an error", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}, nil)

				service.SetupReturns(api.SetupOutput{}, errors.New("could not setup"))

				err := executeCommand(command,commandLineArgs)
				Expect(err).To(MatchError("could not configure authentication: could not setup"))
			})
		})

		When("the final configuration status cannot be determined", func() {
			It("returns an error", func() {
				eaOutputs := []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusUnstarted},
				}
				eaErrors := []error{nil, nil, nil, errors.New("failed to fetch status")}

				service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
					return eaOutputs[service.EnsureAvailabilityCallCount()-1], eaErrors[service.EnsureAvailabilityCallCount()-1]
				}

				err := executeCommand(command,commandLineArgs)
				Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
			})
		})

		When("missing required fields", func() {
			It("returns an error", func() {
				command := commands.NewConfigureLDAPAuthentication(nil, nil, nil)
				err := executeCommand(command,nil)
				Expect(err).To(MatchError("could not parse configure-ldap-authentication flags: missing required flag \"--decryption-passphrase\""))
			})
		})

	})
})
