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

var _ = Describe("CredentialReferences", func() {
	var (
		fakeService   *fakes.CredentialReferencesService
		fakePresenter *presenterfakes.FormattedPresenter
		logger        *fakes.Logger

		command commands.CredentialReferences
	)

	BeforeEach(func() {
		fakeService = &fakes.CredentialReferencesService{}
		fakePresenter = &presenterfakes.FormattedPresenter{}
		logger = &fakes.Logger{}

		command = commands.NewCredentialReferences(fakeService, fakePresenter, logger)
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "other-deployed-product-guid",
				}}, nil)

			fakeService.ListDeployedProductCredentialsReturns(api.CredentialReferencesOutput{
				Credentials: []string{
					".properties.some-credentials",
					".our-job.some-other-credential",
					".my-job.some-credentials",
				},
			}, nil)
		})

		It("lists the credential references in alphabetical order", func() {
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePresenter.PresentCredentialReferencesCallCount()).To(Equal(1))
			Expect(fakePresenter.PresentCredentialReferencesArgsForCall(0)).To(ConsistOf(
				".my-job.some-credentials",
				".our-job.some-other-credential",
				".properties.some-credentials",
			))
		})

		When("the format flag is provided", func() {
			It("sets format on the presenter", func() {
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--format", "json",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePresenter.SetFormatCallCount()).To(Equal(1))
				Expect(fakePresenter.SetFormatArgsForCall(0)).To(Equal("json"))
			})
		})

		Context("failure cases", func() {
			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse credential-references flags: flag provided but not defined: -badflag"))
				})
			})

			When("the product-name flag is not provided", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse credential-references flags: missing required flag \"--product-name\""))
				})
			})

			When("the deployed product cannot be found", func() {
				BeforeEach(func() {
					fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
				})

				It("returns an error", func() {
					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references")))
				})
			})

			When("there are no credential references to list", func() {
				It("prints a helpful message instead of a table", func() {
					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)

					fakeService.ListDeployedProductCredentialsReturns(api.CredentialReferencesOutput{}, nil)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(fakePresenter.PresentCredentialReferencesCallCount()).To(Equal(0))

					Expect(logger.PrintfArgsForCall(0)).To(Equal("no credential references found"))
				})
			})

			When("the credential references cannot be fetched", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)

					fakeService.ListDeployedProductCredentialsReturns(api.CredentialReferencesOutput{}, errors.New("could not fetch credential references"))

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references: could not fetch credential references")))

					Expect(fakePresenter.PresentCredentialReferencesCallCount()).To(Equal(0))
				})
			})

			When("the deployed products cannot be fetched", func() {
				It("returns an error", func() {
					fakeService.ListDeployedProductsReturns(
						[]api.DeployedProductOutput{},
						errors.New("could not fetch deployed products"))

					command := commands.NewCredentialReferences(fakeService, fakePresenter, logger)
					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references: could not fetch deployed products")))

					Expect(fakePresenter.PresentCredentialReferencesCallCount()).To(Equal(0))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewCredentialReferences(nil, nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command lists credential references for deployed products.",
				ShortDescription: "list credential references for a deployed product",
				Flags:            command.Options,
			}))
		})
	})
})
