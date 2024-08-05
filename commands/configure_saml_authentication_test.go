package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureSAMLAuthentication.Execute", func() {
	var (
		service         *fakes.ConfigureAuthenticationService
		logger          *fakes.Logger
		command         *commands.ConfigureSAMLAuthentication
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

		err := executeCommand(command, commandLineArgs)
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
			err := executeCommand(command, commandLineArgs)
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
			err := executeCommand(command, commandLineArgs)
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

			err := executeCommand(command, commandLineArgs)
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
				err := executeCommand(command, commandLineArgs)
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
			err := executeCommand(command, commandLineArgs)
			Expect(err).ToNot(HaveOccurred())

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
			Expect(service.SetupCallCount()).To(Equal(0))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuration previously completed, skipping configuration"))
		})
	})

	When("the initial configuration status cannot be determined", func() {
		It("returns an error", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

			command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
			err := executeCommand(command, commandLineArgs)
			Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
		})
	})

	When("the initial configuration status is unknown", func() {
		It("returns an error", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusUnknown,
			}, nil)

			command := commands.NewConfigureSAMLAuthentication(nil, service, &fakes.Logger{})
			err := executeCommand(command, commandLineArgs)
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
			err := executeCommand(command, commandLineArgs)
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
			err := executeCommand(command, commandLineArgs)
			Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
		})
	})
})
