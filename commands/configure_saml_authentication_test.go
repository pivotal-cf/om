package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("ConfigureSAMLAuthentication.Execute", func() {
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

		command = commands.NewConfigureSAMLAuthentication(nil, service, logger)

		commandLineArgs = []string{
			"--decryption-passphrase", "some-passphrase",
			"--saml-idp-metadata", "https://saml.example.com:8080",
			"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
			"--saml-rbac-admin-group", "opsman.full_control",
			"--saml-rbac-groups-attribute", "myenterprise",
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
			CreateBoshAdminClient:            true,
		}
	})

	It("configures SAML authentication", func() {
		commandLineArgs = append(commandLineArgs, "--precreated-client-secret", "test-client-secret")
		expectedPayload.PrecreatedClientSecret = "test-client-secret"

		err := command.Execute(commandLineArgs)
		Expect(err).ToNot(HaveOccurred())

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

			expectedPayload.CreateBoshAdminClient = false
		})

		It("configures SAML with bosh admin client warning", func() {
			err := command.Execute(commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

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

			commandLineArgs = append(commandLineArgs, "--precreated-client-secret", "test-client-secret")
		})

		It("errors out if you try to provide a client secret", func() {
			err := command.Execute(commandLineArgs)
			Expect(err).To(MatchError(ContainSubstring(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`)))
		})
	})

	When("the skip-create-bosh-admin-client flag is set", func() {
		BeforeEach(func() {
			expectedPayload.CreateBoshAdminClient = false
			commandLineArgs = append(commandLineArgs, "--skip-create-bosh-admin-client")
		})

		It("configures SAML auth and notifies the user that it skipped client creation", func() {
			expectedPayload.PrecreatedClientSecret = "test-client-secret"
			commandLineArgs = append(commandLineArgs, "--precreated-client-secret", "test-client-secret")

			err := command.Execute(commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

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
				expectedPayload.CreateBoshAdminClient = false
			})

			It("configures SAML and notifies the user that it skipped client creation", func() {
				err := command.Execute(commandLineArgs)
				Expect(err).ToNot(HaveOccurred())

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

	When("the authentication setup has already been configured", func() {
		It("returns without configuring the authentication system", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusComplete,
			}, nil)

			command := commands.NewConfigureSAMLAuthentication(nil, service, logger)
			err := command.Execute(commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
			Expect(service.SetupCallCount()).To(Equal(0))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuration previously completed, skipping configuration"))
		})
	})

	When("a complete config file is provided", func() {
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
			Expect(err).ToNot(HaveOccurred())

			_, err = configFile.WriteString(configContent)
			Expect(err).ToNot(HaveOccurred())

			expectedPayload.PrecreatedClientSecret = "test-client-secret"
		})

		It("reads configuration from config file", func() {
			err := command.Execute([]string{
				"--config", configFile.Name(),
			})
			Expect(err).ToNot(HaveOccurred())

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
			Expect(err).ToNot(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
				IdentityProvider:                 "saml",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase",
				EULAAccepted:                     "true",
				IDPMetadata:                      "https://super.example.com:6543",
				BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
				RBACAdminGroup:                   "opsman.full_control",
				RBACGroupsAttribute:              "myenterprise",
				CreateBoshAdminClient:            true,
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

	When("the config file contains variables", func() {
		var configFile string

		BeforeEach(func() {
			service.EnsureAvailabilityReturnsOnCall(0, api.EnsureAvailabilityOutput{Status: api.EnsureAvailabilityStatusUnstarted}, nil)
			service.EnsureAvailabilityReturnsOnCall(1, api.EnsureAvailabilityOutput{Status: api.EnsureAvailabilityStatusComplete}, nil)

			configContent := `
decryption-passphrase: ((passphrase))
saml-idp-metadata: https://saml.example.com:8080
saml-bosh-idp-metadata: https://bosh-saml.example.com:8080
saml-rbac-admin-group: opsman.full_control
saml-rbac-groups-attribute: myenterprise
`

			configFile = writeTestConfigFile(configContent)
		})

		Context("variables are not provided", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, service, logger)
				err := command.Execute([]string{
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
`

				file, err := ioutil.TempFile("", "vars-*.yml")
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(file.Name(), []byte(varsContent), 0777)
				Expect(err).ToNot(HaveOccurred())

				varsFile = file.Name()
			})

			It("uses values from the vars file", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, service, logger)
				err := command.Execute([]string{
					"--config", configFile,
					"--vars-file", varsFile,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "saml",
					DecryptionPassphrase:             "a-vars-file-passphrase",
					DecryptionPassphraseConfirmation: "a-vars-file-passphrase",
					EULAAccepted:                     "true",
					IDPMetadata:                      "https://saml.example.com:8080",
					BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
					RBACAdminGroup:                   "opsman.full_control",
					RBACGroupsAttribute:              "myenterprise",
					CreateBoshAdminClient:            true,
				}))
			})
		})

		Context("passed in a var (--var)", func() {
			It("uses values from the command line", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, service, logger)
				err := command.Execute([]string{
					"--config", configFile,
					"--var", "passphrase=a-command-line-passphrase",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "saml",
					DecryptionPassphrase:             "a-command-line-passphrase",
					DecryptionPassphraseConfirmation: "a-command-line-passphrase",
					EULAAccepted:                     "true",
					IDPMetadata:                      "https://saml.example.com:8080",
					BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
					RBACAdminGroup:                   "opsman.full_control",
					RBACGroupsAttribute:              "myenterprise",
					CreateBoshAdminClient:            true,
				}))
			})
		})

		Context("passed as environment variables (--vars-env)", func() {
			It("interpolates variables into the configuration", func() {
				command := commands.NewConfigureSAMLAuthentication(
					func() []string { return []string{"OM_VAR_passphrase=an-env-var-passphrase"} },
					service,
					logger,
				)

				err := command.Execute([]string{
					"--config", configFile,
					"--vars-env", "OM_VAR",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "saml",
					DecryptionPassphrase:             "an-env-var-passphrase",
					DecryptionPassphraseConfirmation: "an-env-var-passphrase",
					EULAAccepted:                     "true",
					IDPMetadata:                      "https://saml.example.com:8080",
					BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
					RBACAdminGroup:                   "opsman.full_control",
					RBACGroupsAttribute:              "myenterprise",
					CreateBoshAdminClient:            true,
				}))
			})

			It("supports the experimental feature of OM_VARS_ENV", func() {
				err := os.Setenv("OM_VARS_ENV", "OM_VAR")
				Expect(err).ToNot(HaveOccurred())
				defer os.Unsetenv("OM_VARS_ENV")

				command := commands.NewConfigureSAMLAuthentication(
					func() []string { return []string{"OM_VAR_passphrase=an-env-var-passphrase"} },
					service,
					logger,
				)

				err = command.Execute([]string{
					"--config", configFile,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "saml",
					DecryptionPassphrase:             "an-env-var-passphrase",
					DecryptionPassphraseConfirmation: "an-env-var-passphrase",
					EULAAccepted:                     "true",
					IDPMetadata:                      "https://saml.example.com:8080",
					BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
					RBACAdminGroup:                   "opsman.full_control",
					RBACGroupsAttribute:              "myenterprise",
					CreateBoshAdminClient:            true,
				}))
			})
		})
	})

	Context("failure cases", func() {
		When("an unknown flag is provided", func() {
			It("returns an error", func() {
				err := command.Execute([]string{"--banana"})
				Expect(err).To(MatchError("could not parse configure-saml-authentication flags: flag provided but not defined: -banana"))
			})
		})

		When("config file cannot be opened", func() {
			It("returns an error", func() {
				err := command.Execute([]string{"--config", "something"})
				Expect(err).To(MatchError("could not parse configure-saml-authentication flags: could not load the config file: could not read file (something): open something: no such file or directory"))

			})
		})

		When("the initial configuration status cannot be determined", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

				command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
				err := command.Execute(commandLineArgs)
				Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
			})
		})

		When("the initial configuration status is unknown", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnknown,
				}, nil)

				command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
				err := command.Execute(commandLineArgs)
				Expect(err).To(MatchError("could not determine initial configuration status: received unexpected status"))
			})
		})

		When("the setup service encounters an error", func() {
			It("returns an error", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}, nil)

				service.SetupReturns(api.SetupOutput{}, errors.New("could not setup"))

				command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
				err := command.Execute(commandLineArgs)
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

				command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
				err := command.Execute(commandLineArgs)
				Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
			})
		})

		When("the --saml-idp-metadata field is not configured with others", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, nil, nil)
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

		When("the --saml-bosh-idp-metadata field is not configured with others", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, nil, nil)
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

		When("the --saml-rbac-admin-group field is not configured with others", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, nil, nil)
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

		When("the --saml-rbac-groups-attribute field is not configured with others", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, nil, nil)
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

		When("the --decryption-passphrase flag is missing", func() {
			It("returns an error", func() {
				command := commands.NewConfigureSAMLAuthentication(nil, nil, nil)
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
