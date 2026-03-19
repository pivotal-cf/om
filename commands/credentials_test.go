package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials", func() {
	var (
		fakeService   *fakes.CredentialsService
		fakePresenter *presenterfakes.FormattedPresenter
		logger        *fakes.Logger

		command *commands.Credentials
	)

	BeforeEach(func() {
		fakeService = &fakes.CredentialsService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		logger = &fakes.Logger{}

		command = commands.NewCredentials(fakeService, fakePresenter, logger)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "some-deployed-product-guid",
				}}, nil)

			fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
				Credential: api.Credential{
					Type: "simple_credentials",
					Value: map[string]interface{}{
						"password": "some-password",
						"identity": "some-identity",
					},
				},
			}, nil)
		})

		Describe("outputting all values for a credential", func() {
			It("outputs the credentials alphabetically", func() {
				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--credential-reference", ".properties.some-credentials",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeService.GetDeployedProductCredentialCallCount()).To(Equal(1))
				Expect(fakeService.GetDeployedProductCredentialArgsForCall(0)).To(Equal(api.GetDeployedProductCredentialInput{
					DeployedGUID:        "some-deployed-product-guid",
					CredentialReference: ".properties.some-credentials",
				}))

				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))

				Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentCredentialsArgsForCall(0)).To(Equal(
					map[string]interface{}{
						"password": "some-password",
						"identity": "some-identity",
					},
				))
			})

			When("the format flag is provided", func() {
				It("sets the format on the presenter", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
						"--format", "json",
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
				})
			})

			When("the credential reference cannot be found", func() {
				BeforeEach(func() {
					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, nil)
				})

				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential for "some-credential"`)))
				})
			})

			When("the credentials cannot be fetched", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, errors.New("could not fetch credentials"))

					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential for "some-credential": could not fetch credentials`)))

					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(0))
				})
			})

			When("the deployed products cannot be fetched", func() {
				It("returns an error", func() {
					fakeService.ListDeployedProductsReturns(
						[]api.DeployedProductOutput{},
						errors.New("could not fetch deployed products"))

					command := commands.NewCredentials(fakeService, fakePresenter, logger)
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to fetch credential: could not fetch deployed products")))

					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(0))
				})
			})

			When("the product has not been deployed", func() {
				It("returns an error", func() {
					fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
						api.DeployedProductOutput{
							Type: "some-other-product",
							GUID: "some-other-deployed-product-guid",
						}}, nil)

					command := commands.NewCredentials(fakeService, fakePresenter, logger)
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential: "some-product" is not deployed`)))

					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(0))
				})
			})
		})

		Describe("outputting an individual credential value", func() {
			BeforeEach(func() {
				fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
					Credential: api.Credential{
						Type: "simple_credentials",
						Value: map[string]interface{}{
							"password": "some-password",
							"identity": "some-identity",
						},
					},
				}, nil)
			})

			It("outputs the credential value only", func() {
				command := commands.NewCredentials(fakeService, fakePresenter, logger)

				err := executeCommand(command, []string{
					"--product-name", "some-product",
					"--credential-reference", ".properties.some-credentials",
					"--credential-field", "password",
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(logger.PrintlnCallCount()).To(Equal(1))
				Expect(logger.PrintlnArgsForCall(0)[0]).To(Equal("some-password"))
			})

			When("the credential field cannot be found", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
						"--credential-field", "missing-field",
					})
					Expect(err).To(MatchError(ContainSubstring(`credential field "missing-field" not found`)))
				})
			})
		})

		Describe("generated_credentials", func() {
			When("output contains a nested hash value", func() {
				BeforeEach(func() {
					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
						Credential: api.Credential{
							Type: "generated_credentials",
							Value: map[string]interface{}{
								"output": map[string]interface{}{
									"host": "10.0.0.1",
									"port": float64(8080),
									"username": "admin",
								},
							},
						},
					}, nil)
				})

				It("presents all credential values including the output key", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))

					presented := fakePresenter.PresentCredentialsArgsForCall(0)
					valMap, ok := presented.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(valMap).To(HaveKey("output"))
				})

				It("presents the complex field value via the presenter when using --credential-field output", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
						"--credential-field", "output",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
					Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))
					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))

					presented := fakePresenter.PresentCredentialsArgsForCall(0)
					valMap, ok := presented.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(valMap).To(HaveKeyWithValue("host", "10.0.0.1"))
					Expect(valMap).To(HaveKeyWithValue("username", "admin"))
				})

				It("respects --format json when using --credential-field output on a complex value", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
						"--credential-field", "output",
						"--format", "json",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
					Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				})

				It("returns an error for a field that does not exist", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
						"--credential-field", "nonexistent",
					})
					Expect(err).To(MatchError(ContainSubstring(`credential field "nonexistent" not found`)))
				})
			})

			When("output contains a simple string value", func() {
				BeforeEach(func() {
					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
						Credential: api.Credential{
							Type: "generated_credentials",
							Value: map[string]interface{}{
								"output": "some-simple-generated-string",
							},
						},
					}, nil)
				})

				It("presents the value including the output key", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				})

				It("outputs the string when using --credential-field output", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
						"--credential-field", "output",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(logger.PrintlnCallCount()).To(Equal(1))
					Expect(logger.PrintlnArgsForCall(0)[0]).To(Equal("some-simple-generated-string"))
				})
			})

			When("output is empty (credentials not yet generated)", func() {
				BeforeEach(func() {
					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
						Credential: api.Credential{
							Type: "generated_credentials",
							Value: map[string]interface{}{
								"output": "",
							},
						},
					}, nil)
				})

				It("presents the value without error", func() {
					err := executeCommand(command, []string{
						"--product-name", "some-product",
						"--credential-reference", ".properties.generated-cred",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				})
			})
		})
	})
})
