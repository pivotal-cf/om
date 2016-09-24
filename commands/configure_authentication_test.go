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
		It("sets up a user with the specified configuration information", func() {
			service := &fakes.SetupService{}

			command := commands.NewConfigureAuthentication(service)
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
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(&fakes.SetupService{})
					err := command.Execute([]string{"--banana"})
					Expect(err).To(MatchError("could not parse configure-authentication flags: flag provided but not defined: -banana"))
				})
			})

			Context("when the setup service encounters an error", func() {
				It("returns an error", func() {
					service := &fakes.SetupService{}
					service.SetupCall.Returns.Error = errors.New("could not setup")

					command := commands.NewConfigureAuthentication(service)
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not configure authentication: could not setup"))
				})
			})
		})
	})

	Describe("Help", func() {
		It("returns a short help description of the command", func() {
			command := commands.NewConfigureAuthentication(nil)
			Expect(command.Help()).To(Equal("configures OpsManager with an internal userstore and admin user account"))
		})
	})
})
