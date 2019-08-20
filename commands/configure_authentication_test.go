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

var _ = Describe("ConfigureAuthentication", func() {
	var (
		logger  *fakes.Logger
		service *fakes.ConfigureAuthenticationService
	)

	BeforeEach(func() {
		service = &fakes.ConfigureAuthenticationService{}
		service.InfoReturns(api.Info{
			Version: "2.5-build.1",
		}, nil)
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
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

			command := commands.NewConfigureAuthentication(service, logger)
			err := command.Execute([]string{
				"--username", "some-username",
				"--password", "some-password",
				"--decryption-passphrase", "some-passphrase",
				"--precreated-client-secret", "test-client-secret",
			})
			Expect(err).NotTo(HaveOccurred())

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

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring internal userstore..."))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal(`
Ops Manager UAA client will be created when authentication system starts.
It will have the username 'precreated-client' and the client secret you provided.
`))
		})

		When("the authentication setup has already been configured", func() {
			It("returns without configuring the authentication system", func() {
				service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}, nil)

				command := commands.NewConfigureAuthentication(service, logger)
				err := command.Execute([]string{
					"--username", "some-username",
					"--password", "some-password",
					"--decryption-passphrase", "some-passphrase",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(1))
				Expect(service.SetupCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration previously completed, skipping configuration"))
			})
		})

		When("config file is provided", func() {
			var configFile *os.File

			BeforeEach(func() {
				var err error
				configContent := `
username: some-username
password: some-password
decryption-passphrase: some-passphrase
`
				configFile, err = ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.WriteString(configContent)
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads configuration from config file", func() {
				eaOutputs := []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusComplete},
				}

				service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
					return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
				}

				command := commands.NewConfigureAuthentication(service, logger)
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "internal",
					AdminUserName:                    "some-username",
					AdminPassword:                    "some-password",
					AdminPasswordConfirmation:        "some-password",
					DecryptionPassphrase:             "some-passphrase",
					DecryptionPassphraseConfirmation: "some-passphrase",
					EULAAccepted:                     "true",
				}))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring internal userstore..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
			})

			It("is overridden by commandline flags", func() {
				eaOutputs := []api.EnsureAvailabilityOutput{
					{Status: api.EnsureAvailabilityStatusUnstarted},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusPending},
					{Status: api.EnsureAvailabilityStatusComplete},
				}

				service.EnsureAvailabilityStub = func(api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
					return eaOutputs[service.EnsureAvailabilityCallCount()-1], nil
				}

				command := commands.NewConfigureAuthentication(service, logger)
				err := command.Execute([]string{
					"--config", configFile.Name(),
					"--password", "some-password-1",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.SetupArgsForCall(0)).To(Equal(api.SetupInput{
					IdentityProvider:                 "internal",
					AdminUserName:                    "some-username",
					AdminPassword:                    "some-password-1",
					AdminPasswordConfirmation:        "some-password-1",
					DecryptionPassphrase:             "some-passphrase",
					DecryptionPassphraseConfirmation: "some-passphrase",
					EULAAccepted:                     "true",
				}))

				Expect(service.EnsureAvailabilityCallCount()).To(Equal(4))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring internal userstore..."))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("waiting for configuration to complete..."))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuration complete"))
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

				command := commands.NewConfigureAuthentication(service, logger)
				err := command.Execute([]string{
					"--username", "some-username",
					"--password", "some-password",
					"--decryption-passphrase", "some-passphrase",
					"--precreated-client-secret", "test-client-secret",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`
Cannot use the "--precreated-client-secret" argument.
This is only supported in OpsManager 2.5 and up.
`))
			})
		})

		Context("failure cases", func() {
			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{"--banana"})
					Expect(err).To(MatchError("could not parse configure-authentication flags: flag provided but not defined: -banana"))
				})
			})

			When("config file cannot be opened", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{"--config", "something"})
					Expect(err).To(MatchError("could not parse configure-authentication flags: could not load the config file: could not read file (something): open something: no such file or directory"))

				})
			})

			When("the initial configuration status cannot be determined", func() {
				It("returns an error", func() {
					service.EnsureAvailabilityReturns(api.EnsureAvailabilityOutput{}, errors.New("failed to fetch status"))

					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{
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

					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{
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

					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{
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

					command := commands.NewConfigureAuthentication(service, logger)
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not determine final configuration status: failed to fetch status"))
				})
			})

			When("the --username flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(nil, nil)
					err := command.Execute([]string{
						"--password", "some-password",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not parse configure-authentication flags: missing required flag \"--username\""))
				})
			})

			When("the --password flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(nil, nil)
					err := command.Execute([]string{
						"--username", "some-username",
						"--decryption-passphrase", "some-passphrase",
					})
					Expect(err).To(MatchError("could not parse configure-authentication flags: missing required flag \"--password\""))
				})
			})

			When("the --decryption-passphrase flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewConfigureAuthentication(nil, nil)
					err := command.Execute([]string{
						"--username", "some-username",
						"--password", "some-password",
					})
					Expect(err).To(MatchError("could not parse configure-authentication flags: missing required flag \"--decryption-passphrase\""))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureAuthentication(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This unauthenticated command helps setup the internal userstore authentication mechanism for your Ops Manager.",
				ShortDescription: "configures Ops Manager with an internal userstore and admin user account",
				Flags:            command.Options,
			}))
		})
	})
})
