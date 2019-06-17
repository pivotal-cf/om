package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("ConfigureSAMLAuthentication", func() {
	Describe("Execute", func() {
		var (
			service         *fakes.ConfigureAuthenticationService
			logger          *fakes.Logger
			command         commands.ConfigureSAMLAuthentication
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

			service.InfoReturns(api.Info{
				Version: "2.5-build.1",
			}, nil)

			command = commands.NewConfigureSAMLAuthentication(service, logger)

			commandLineArgs = []string{
				"--decryption-passphrase", "some-passphrase",
				"--saml-idp-metadata", "https://saml.example.com:8080",
				"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
				"--saml-rbac-admin-group", "opsman.full_control",
				"--saml-rbac-groups-attribute", "myenterprise",
				"--precreated-client-secret", "test-client-secret",
			}

			expectedPayload = api.SetupInput{
				IdentityProvider:                 "saml",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase",
				EULAAccepted:                     "true",
				IDPMetadata:                      "https://saml.example.com:8080",
				BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
				RBACAdminGroup:                   "opsman.full_control",
				RBACGroupsAttribute:              "myenterprise",
				CreateBoshAdminClient:            "true",
				PrecreatedClientSecret:           "test-client-secret",
			}
		})

		It("configures SAML authentication", func() {
			err := command.Execute(commandLineArgs)
			Expect(err).NotTo(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

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

				expectedPayload.CreateBoshAdminClient = ""
				expectedPayload.PrecreatedClientSecret = ""
			})

			It("configures SAML with bosh admin client warning", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

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

				expectedPayload.PrecreatedClientSecret = ""
			})

			It("configures SAML with OpsMan UAA client warning", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

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
Note: Ops Manager UAA client NOT automatically created.
This is only supported in OpsManager 2.5 and up.
`))
			})
		})

		When("the skip-create-bosh-admin-client flag is set", func() {
			BeforeEach(func() {
				commandLineArgs = append(commandLineArgs, "--skip-create-bosh-admin-client")
				expectedPayload.CreateBoshAdminClient = "false"
			})

			It("configures SAML auth and notifies the user that it skipped client creation", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

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
					expectedPayload.CreateBoshAdminClient = ""
					expectedPayload.PrecreatedClientSecret = ""
				})
				It("configures SAML and notifies the user that it skipped client creation", func() {
					err := command.Execute(commandLineArgs)
					Expect(err).NotTo(HaveOccurred())

					Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

					Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

					format, content := logger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

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
			It("returns without configuring the authentication system", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}, nil)

				command := commands.NewConfigureSAMLAuthentication(service, logger)
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
saml-idp-metadata: https://saml.example.com:8080
saml-bosh-idp-metadata: https://bosh-saml.example.com:8080
saml-rbac-admin-group: opsman.full_control
saml-rbac-groups-attribute: myenterprise
decryption-passphrase: some-passphrase
precreated-client-secret: test-client-secret
`
				configFile, err = ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(configContent)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads configuration from config file", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(expectedPayload))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
			})

			It("is overridden by commandline flags", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
					"--saml-idp-metadata", "https://super.example.com:6543",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "saml",
					DecryptionPassphrase:             "some-passphrase",
					DecryptionPassphraseConfirmation: "some-passphrase",
					EULAAccepted:                     "true",
					IDPMetadata:                      "https://super.example.com:6543",
					BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
					RBACAdminGroup:                   "opsman.full_control",
					RBACGroupsAttribute:              "myenterprise",
					CreateBoshAdminClient:            "true",
					PrecreatedClientSecret:           "test-client-secret",
				}))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
			})
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--banana"})
					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: flag provided but not defined: -banana"))
				})
			})

			Context("when config file cannot be opened", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--config", "something"})
					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: could not load the config file: open something: no such file or directory"))

				})
			})

			Context("when the initial configuration status cannot be determined", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
				})
			})

			Context("when the initial configuration status is unknown", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
						Status: api.EnsureAvailabilityStatusUnknown,
					}, nil)

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
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

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
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

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
					err := command.Execute(commandLineArgs)
					Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
				})
			})

			Context("when the --saml-idp-metadata field is not configured with others", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(nil, nil)
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(HaveOccurred())

					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: missing required flag \"--saml-idp-metadata\""))
				})
			})

			Context("when the --saml-bosh-idp-metadata field is not configured with others", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(nil, nil)
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(HaveOccurred())

					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: missing required flag \"--saml-bosh-idp-metadata\""))
				})
			})

			Context("when the --saml-rbac-admin-group field is not configured with others", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(nil, nil)
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(HaveOccurred())

					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: missing required flag \"--saml-rbac-admin-group\""))
				})
			})

			Context("when the --saml-rbac-groups-attribute field is not configured with others", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(nil, nil)
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
					})
					Expect(err).To(HaveOccurred())

					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: missing required flag \"--saml-rbac-groups-attribute\""))
				})
			})

			Context("when the --decryption-passphrase flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(nil, nil)
					err := command.Execute([]string{
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
					})
					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: missing required flag \"--decryption-passphrase\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureSAMLAuthentication(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This unauthenticated command helps setup the authentication mechanism for your Ops Manager with SAML.",
				ShortDescription: "configures Ops Manager with SAML authentication",
				Flags:            command.Options,
			}))
		})
	})
})
