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
		csService     *fakes.CredentialsService
		dpLister      *fakes.DeployedProductsLister
		fakePresenter *presenterfakes.Presenter
		logger        *fakes.Logger
	)

	BeforeEach(func() {
		csService = &fakes.CredentialsService{}
		dpLister = &fakes.DeployedProductsLister{}
		fakePresenter = &presenterfakes.Presenter{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			dpLister.ListReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "some-deployed-product-guid",
				}}, nil)
		})

		Describe("outputting all values for a credential", func() {
			It("outputs the credentials alphabetically", func() {
				command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

				csService.FetchCredentialReturns(api.CredentialOutput{
					Credential: api.Credential{
						Type: "simple_credentials",
						Value: map[string]string{
							"password": "some-password",
							"identity": "some-identity",
						},
					},
				}, nil)

				err := command.Execute([]string{
					"--product-name", "some-product",
					"--credential-reference", ".properties.some-credentials",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakePresenter.PresentCredentialsCallCount()).To(Equal(1))
				Expect(fakePresenter.PresentCredentialsArgsForCall(0)).To(Equal(

					map[string]string{
						"password": "some-password",
						"identity": "some-identity",
					},
				))
			})

			Context("when the --product-name flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

					err := command.Execute([]string{
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError("could not parse credential-references flags: missing required flag \"--product-name\""))
				})
			})

			Context("when the --credential-reference flag is missing", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError("could not parse credential-references flags: missing required flag \"--credential-reference\""))
				})
			})

			Context("when the credential reference cannot be found", func() {
				BeforeEach(func() {
					csService.FetchCredentialReturns(api.CredentialOutput{}, nil)
				})

				It("returns an error", func() {
					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
						"--credential-reference", "some-credential",
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to fetch credential for "some-credential"`)))
				})
			})

			Context("when the credentials cannot be fetched", func() {
				It("returns an error", func() {
					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

					csService.FetchCredentialReturns(api.CredentialOutput{}, errors.New("could not fetch credentials"))

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
					dpLister.ListReturns(
						[]api.DeployedProductOutput{},
						errors.New("could not fetch deployed products"))

					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)
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
					dpLister.ListReturns([]api.DeployedProductOutput{
						api.DeployedProductOutput{
							Type: "some-other-product",
							GUID: "some-other-deployed-product-guid",
						}}, nil)

					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)
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
				csService.FetchCredentialReturns(api.CredentialOutput{
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
				command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

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
					command := commands.NewCredentials(csService, dpLister, fakePresenter, logger)

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
			command := commands.NewCredentials(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command fetches credentials for deployed products.",
				ShortDescription: "fetch credentials for a deployed product",
				Flags:            command.Options,
			}))
		})
	})
})
