package commands_test

import (
	"errors"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureAuthentication.Execute", func() {
	var (
		stdout  *gbytes.Buffer
		logger  *log.Logger
		service *fakes.ConfigureAuthenticationService
	)

	BeforeEach(func() {
		service = &fakes.ConfigureAuthenticationService{}
		service.InfoReturns(api.Info{
			Version: "2.5-build.1",
		}, nil)
		stdout = gbytes.NewBuffer()
		logger = log.New(stdout, "", 0)
	})

	It("sets up a user with the specified configuration information, waiting for the setup to complete", func() {
		eaOutputs := []api.EnsureAvailabilityOutput{
			{Status: api.EnsureAvailabilityStatusUnstarted},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusPending},
			{Status: api.EnsureAvailabilityStatusComplete},
		}

		service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
			return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
		}

		command := commands.NewConfigureAuthentication(nil, service, logger)
		err := executeCommand(command, []string{
			"--username", "some-username",
			"--password", "some-password",
			"--decryption-passphrase", "some-passphrase",
			"--precreated-client-secret", "test-client-secret",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
			IdentityProvider:                 "internal",
			AdminUserName:                    "some-username",
			AdminPassword:                    "some-password",
			AdminPasswordConfirmation:        "some-password",
			DecryptionPassphrase:             "some-passphrase",
			DecryptionPassphraseConfirmation: "some-passphrase",
			EULAAccepted:                     "true",
			PrecreatedClientSecret:           "test-client-secret",
		}))

		Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

		Expect(stdout).To(gbytes.Say("configuring internal userstore..."))
		Expect(stdout).To(gbytes.Say("waiting for configuration to complete..."))
		Expect(stdout).To(gbytes.Say("configuration complete"))
		Expect(stdout).To(gbytes.Say(`
Ops Manager UAA client will be created when authentication system starts.
It will have the username 'precreated-client' and the client secret you provided.
`))
	})

	When("the authentication setup has already been configured", func() {
		It("returns without configuring the authentication system", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusComplete,
			}, nil)

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
			Expect(service.SetupCallCount()).To(Equal(0))

			Expect(stdout).To(gbytes.Say("configuration previously completed, skipping configuration"))
		})
	})

	When("OpsMan is < 2.5", func() {
		It("errors out if you try to provide a client secret", func() {
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
				Version: "2.4-build.1",
			}, nil)

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
				"--precreated-client-secret", "test-client-secret",
			})
			Expect(err).To(MatchError(ContainSubstring(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`)))
		})
	})

	When("the initial configuration status cannot be determined", func() {
		It("returns an error", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
		})
	})

	When("the initial configuration status is unknown", func() {
		It("returns an error", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusUnknown,
			}, nil)

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).To(MatchError("could not determine initial configuration status: received unexpected status"))
		})
	})

	When("the setup service encounters an error", func() {
		It("returns an error", func() {
			service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
				Status: api.EnsureAvailabilityStatusUnstarted,
			}, nil)

			service.SetupReturns(api.SetupOutput{}, errors.New("could not setup"))

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
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

			command := commands.NewConfigureAuthentication(nil, service, logger)
			err := executeCommand(command, []string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
		})
	})
})
