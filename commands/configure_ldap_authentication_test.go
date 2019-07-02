package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureLDAPAuthentication", func() {
	Describe("Execute", func() {
		var (
			service         *fakes.ConfigureAuthenticationService
			logger          *fakes.Logger
			command         commands.ConfigureLDAPAuthentication
			commandLineArgs []string
			expectedPayload api.SetupInput
		)

		BeforeEach(func() {
			service = &fakes.ConfigureAuthenticationService{}
			logger = &fakes.Logger{}

			eaOutputs := []api.EnsureAvailabilityOutput{
				{Status: api.EnsureAvailabilityStatusUnstarted},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusComplete},
			}
			service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
				return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
			}

			command = commands.NewConfigureLDAPAuthentication(service, logger)

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

			err := command.Execute(commandLineArgs)
			Expect(err).NotTo(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring LDAP authentication..."))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal(`
BOSH admin client will be created when the director is deployed.
The client secret can then be found in the Ops Manager UI:
director tile -> Credentials tab -> click on 'Link to Credential' for 'Uaa Bosh Client Credentials'
Note both the client ID and secret.
`))

			format, content = logger.PrintfArgsForCall(4)
			Expect(fmt.Sprintf(format, content...)).To(Equal(`
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
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring LDAP authentication..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal(`
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
				err := command.Execute(commandLineArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`))
			})
		})

		When("the skip-create-bosh-admin-client flag is set", func() {
			BeforeEach(func() {
				commandLineArgs = append(commandLineArgs, "--skip-create-bosh-admin-client")
				expectedPayload.CreateBoshAdminClient = false
			})

			It("configures LDAP auth and notifies the user that it skipped client creation", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring LDAP authentication..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal(`
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
					err := command.Execute(commandLineArgs)
					Expect(err).NotTo(HaveOccurred())

					Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

					Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

					format, content := logger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("configuring LDAP authentication..."))

					format, content = logger.PrintfArgsForCall(1)
					Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

					format, content = logger.PrintfArgsForCall(2)
					Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))

					format, content = logger.PrintfArgsForCall(3)
					Expect(fmt.Sprintf(format, content...)).To(Equal(`
Note: BOSH admin client NOT automatically created.
This was skipped due to the 'skip-create-bosh-admin-client' flag.
`))
				})
			})
		})

		Context("when the authentication setup has already been configured", func() {
			BeforeEach(func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}, nil)
			})

			It("returns without configuring the authentication system", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
				Expect(service.SetupCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration previously completed, skipping configuration"))
			})
		})

		Context("when config file is provided", func() {
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
				Expect(err).NotTo(HaveOccurred())
				defer configFile.Close()

				_, err = configFile.WriteString(configContent)
				Expect(err).NotTo(HaveOccurred())
				expectedPayload.PrecreatedClientSecret = "test-client-secret"
			})

			It("reads configuration from config file", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring LDAP authentication..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
			})

			It("is overridden by commandline flags", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
					"--server-url", "ldap://example.com",
				})
				Expect(err).NotTo(HaveOccurred())

				expectedPayload.LDAPSettings.ServerURL = "ldap://example.com"

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))
			})
		})

		Context("failure cases", func() {
			Context("when config file cannot be opened", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--config", "something"})
					Expect(err).To(MatchError("could not parse configure-ldap-authentication flags: could not load the config file: open something: no such file or directory"))

				})
			})

			Context("when the initial configuration status cannot be determined", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
				})
			})

			Context("when the initial configuration status is unknown", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
						Status: api.EnsureAvailabilityStatusUnknown,
					}, nil)

					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not determine initial configuration status: received unexpected status"))
				})
			})

			Context("when the setup service encounters an error", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
						Status: api.EnsureAvailabilityStatusUnstarted,
					}, nil)

					service.SetupReturns(api.SetupOutput{}, errors.New("could not setup"))

					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not configure authentication: could not setup"))
				})
			})

			Context("when the final configuration status cannot be determined", func() {
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

					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
				})
			})

			When("missing required fields", func() {
				It("returns an error", func() {
					command := commands.NewConfigureLDAPAuthentication(nil, nil)
					err := command.Execute(nil)
					Expect(err).To(MatchError("could not parse configure-ldap-authentication flags: missing required flag \"--decryption-passphrase\""))
				})
			})

		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureLDAPAuthentication(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager with LDAP.",
				ShortDescription: "configures Ops Manager with LDAP authentication",
				Flags:            command.Options,
			}))
		})
	})
})
