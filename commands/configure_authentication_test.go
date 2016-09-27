package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureAuthentication", func() {
	Describe("Execute", func() {
		It("sets up a user with the specified configuration information, waiting for the setup to complete", func() {
			service := &fakes.SetupService{}
			service.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
				{Status: api.EnsureAvailabilityStatusUnstarted},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusComplete},
			}
			logger := &fakes.Logger{}

			command := commands.NewConfigureAuthentication(service, logger)
			err := command.Execute([]string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.SetupCall.Receives.Input).To(Equal(api.SetupInput{
				IdentityProvider:                 "internal",
				AdminUserName:                    "some-username",
				AdminPassword:                    "some-password",
				AdminPasswordConfirmation:        "some-password",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase",
				EULAAccepted:                     true,
			}))

			Expect(service.EnsureAvailabilityCall.CallCount).To(Equal(4))

			Expect(logger.Lines).To(Equal([]string{
				"configuring internal userstore...",
				"waiting for configuration to complete...",
				"configuration complete",
			}))
		})

		Context("when the authentication setup has already been configured", func() {
			It("returns without configuring the authentication system", func() {
				service := &fakes.SetupService{}
				service.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusComplete},
				}
				logger := &fakes.Logger{}

				command := commands.NewConfigureAuthentication(service, logger)
				err := command.Execute([]string{
					"--username", "some-username",
					"--password", "some-password",
					"--decryption-passphrase", "some-passphrase",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.EnsureAvailabilityCall.CallCount).To(Equal(1))
				Expect(service.SetupCall.CallCount).To(Equal(0))

				Expect(logger.Lines).To(Equal([]string{
					"configuration previously completed, skipping configuration",
				}))
			})
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(&fakes.SetupService{}, &fakes.Logger{})
					err := command.Execute([]string{"--banana"})
					Expect(err).To(MatchError("could not parse configure-authentication flags: flag provided but not defined: -banana"))
				})
			})

			Context("when the initial configuration status cannot be determined", func() {
				It("returns an error", func() {
					service := &fakes.SetupService{}
					service.EnsureAvailabilityCall.Returns.Errors = []error{errors.New("failed to fetch status")}

					command := commands.NewConfigureAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
				})
			})

			Context("when the setup service encounters an error", func() {
				It("returns an error", func() {
					service := &fakes.SetupService{}
					service.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
						{Status: api.EnsureAvailabilityStatusUnstarted},
					}
					service.SetupCall.Returns.Error = errors.New("could not setup")

					command := commands.NewConfigureAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not configure authentication: could not setup"))
				})
			})

			Context("when the final configuration status cannot be determined", func() {
				It("returns an error", func() {
					service := &fakes.SetupService{}
					service.EnsureAvailabilityCall.Returns.Outputs = []api.EnsureAvailabilityOutput{
						{Status: api.EnsureAvailabilityStatusUnstarted},
					}
					service.EnsureAvailabilityCall.Returns.Errors = []error{nil, nil, nil, errors.New("failed to fetch status")}

					command := commands.NewConfigureAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureAuthentication(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command helps setup the authentication mechanism for your OpsManager.\nThe \"internal\" userstore mechanism is the only currently supported option.",
				ShortDescription: "configures OpsManager with an internal userstore and admin user account",
				Flags:            command.Options,
			}))
		})
	})
})
