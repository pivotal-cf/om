package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	presenterfakes "github.com/pivotal-cf/om/presenters/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials", func() {
	var (
		fakeService   *fakes.CredentialsService
		fakePresenter *presenterfakes.FormattedPresenter
		logger        *fakes.Logger

		command commands.Credentials
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
					Value: map[string]string{
						"password": "some-password",
						"identity": "some-identity",
					},
				},
			}, nil)
		})

		Describe("outputting all values for a credential", func() {
			It("outputs the credentials alphabetically", func() {
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--credential-reference", ".properties.some-credentials",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeService.GetDeployedProductCredentialCallCount()).To(Equal(1))
				Expect(fakeService.GetDeployedProductCredentialArgsForCall(0)).To(Equal(api.GetDeployedProductCredentialInput{
					DeployedGUID:        "some-deployed-product-guid",
					CredentialReference: ".properties.some-credentials",
				}))

				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("table"))

				Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentCredentialsArgsForCall(0)).To(Equal(
					map[string]string{
						"password": "some-password",
						"identity": "some-identity",
					},
				))
			})

			Context("when the format flag is provided", func() {
				It("sets the format on the presenter", func() {
					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
						"--format", "json",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
				})
			})

			Context("when the --product-name flag is missing", func() {
				It("returns an error", func() {
					err := command.Execute([]string{
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError("could not parse credential-references flags: missing required flag \"--product-name\""))
				})
			})

			Context("when the --credential-reference flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError("could not parse credential-references flags: missing required flag \"--credential-reference\""))
				})
			})

			Context("when the credential reference cannot be found", func() {
				BeforeEach(func() {
					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, nil)
				})

				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential for "some-credential"`)))
				})
			})

			Context("when the credentials cannot be fetched", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{}, errors.New("could not fetch credentials"))

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential for "some-credential": could not fetch credentials`)))

					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(0))
				})
			})

			Context("when the deployed products cannot be fetched", func() {
				It("returns an error", func() {
					fakeService.ListDeployedProductsReturns(
						[]api.DeployedProductOutput{},
						errors.New("could not fetch deployed products"))

					command := commands.NewCredentials(fakeService, fakePresenter, logger)
					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to fetch credential: could not fetch deployed products")))

					Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(0))
				})
			})

			Context("when the product has not been deployed", func() {
				It("returns an error", func() {
					fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
						api.DeployedProductOutput{
							Type: "some-other-product",
							GUID: "some-other-deployed-product-guid",
						}}, nil)

					command := commands.NewCredentials(fakeService, fakePresenter, logger)
					err := command.Execute([]string{
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
						Value: map[string]string{
							"password": "some-password",
							"identity": "some-identity",
						},
					},
				}, nil)
			})

			It("outputs the credential value only", func() {
				command := commands.NewCredentials(fakeService, fakePresenter, logger)

				err := command.Execute([]string{
					"--product-name", "some-product",
					"--credential-reference", ".properties.some-credentials",
					"--credential-field", "password",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.PrintlnCallCount()).To(Equal(1))
				Expect(logger.PrintlnArgsForCall(0)[0]).To(Equal("some-password"))
			})

			Context("when the credential field cannot be found", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(fakeService, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
						"--credential-field", "missing-field",
					})
					Expect(err).To(MatchError(ContainSubstring(`credential field "missing-field" not found`)))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewCredentials(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command fetches credentials for deployed products.",
				ShortDescription: "fetch credentials for a deployed product",
				Flags:            command.Options,
			}))
		})
	})
})
