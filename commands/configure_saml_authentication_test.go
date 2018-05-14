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
)

var _ = Describe("ConfigureSAMLAuthentication", func() {
	Describe("Execute", func() {
		It("configures SAML authentication", func() {
			service := &fakes.ConfigureAuthenticationService{}
			eaOutputs := []api.EnsureAvailabilityOutput{
				{Status: api.EnsureAvailabilityStatusUnstarted},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusPending},
				{Status: api.EnsureAvailabilityStatusComplete},
			}

			service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
				return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
			}

			logger := &fakes.Logger{}

			command := commands.NewConfigureSAMLAuthentication(service, logger)
			err := command.Execute([]string{
				"--decryption-passphrase", "some-passphrase",
				"--saml-idp-metadata", "https://saml.example.com:8080",
				"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
				"--saml-rbac-admin-group", "opsman.full_control",
				"--saml-rbac-groups-attribute", "myenterprise",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
				IdentityProvider:                 "saml",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase",
				EULAAccepted:                     "true",
				IDPMetadata:                      "https://saml.example.com:8080",
				BoshIDPMetadata:                  "https://bosh-saml.example.com:8080",
				RBACAdminGroup:                   "opsman.full_control",
				RBACGroupsAttribute:              "myenterprise",
			}))

			Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring SAML authentication..."))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
		})

		Context("when the authentication setup has already been configured", func() {
			It("returns without configuring the authentication system", func() {
				service := &fakes.ConfigureAuthenticationService{}
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}, nil)

				logger := &fakes.Logger{}

				command := commands.NewConfigureSAMLAuthentication(service, logger)
				err := command.Execute([]string{
					"--decryption-passphrase", "some-passphrase",
					"--saml-idp-metadata", "https://saml.example.com:8080",
					"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
					"--saml-rbac-admin-group", "opsman.full_control",
					"--saml-rbac-groups-attribute", "myenterprise",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
				Expect(service.SetupCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration previously completed, skipping configuration"))
			})
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureSAMLAuthentication(&fakes.ConfigureAuthenticationService{}, &fakes.Logger{})
					err := command.Execute([]string{"--banana"})
					Expect(err).To(MatchError("could not parse configure-saml-authentication flags: flag provided but not defined: -banana"))
				})
			})

			Context("when the initial configuration status cannot be determined", func() {
				It("returns an error", func() {
					service := &fakes.ConfigureAuthenticationService{}
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(MatchError("could not determine initial configuration status: failed to fetch status"))
				})
			})

			Context("when the initial configuration status is unknown", func() {
				It("returns an error", func() {
					service := &fakes.ConfigureAuthenticationService{}
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
						Status: api.EnsureAvailabilityStatusUnknown,
					}, nil)

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(MatchError("could not determine initial configuration status: received unexpected status"))
				})
			})

			Context("when the setup service encounters an error", func() {
				It("returns an error", func() {
					service := &fakes.ConfigureAuthenticationService{}
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
						Status: api.EnsureAvailabilityStatusUnstarted,
					}, nil)

					service.SetupReturns(api.SetupOutput{}, errors.New("could not setup"))

					command := commands.NewConfigureSAMLAuthentication(service, &fakes.Logger{})
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
					Expect(err).To(MatchError("could not configure authentication: could not setup"))
				})
			})

			Context("when the final configuration status cannot be determined", func() {
				It("returns an error", func() {
					service := &fakes.ConfigureAuthenticationService{}

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
					err := command.Execute([]string{
						"--decryption-passphrase", "some-passphrase",
						"--saml-idp-metadata", "https://saml.example.com:8080",
						"--saml-bosh-idp-metadata", "https://bosh-saml.example.com:8080",
						"--saml-rbac-admin-group", "opsman.full_control",
						"--saml-rbac-groups-attribute", "myenterprise",
					})
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
