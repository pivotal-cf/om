package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialReferences", func() {
	var (
		crService   *fakes.CredentialReferencesService
		dpLister    *fakes.DeployedProductsLister
		tableWriter *fakes.TableWriter
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		crService = &fakes.CredentialReferencesService{}
		dpLister = &fakes.DeployedProductsLister{}
		tableWriter = &fakes.TableWriter{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			dpLister.DeployedProductsReturns([]api.DeployedProductOutput{
				api.DeployedProductOutput{
					Type: "some-product",
					GUID: "other-deployed-product-guid",
				}}, nil)
		})

		It("lists the credential references", func() {
			command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)

			crService.ListReturns(api.CredentialReferencesOutput{
				Credentials: []string{
					".properties.some-credentials",
					".my-job.some-credentials",
				},
			}, nil)

			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Credentials"}))

			Expect(tableWriter.AppendCallCount()).To(Equal(2))
			Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{".properties.some-credentials"}))
			Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{".my-job.some-credentials"}))

			Expect(tableWriter.RenderCallCount()).To(Equal(1))
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse credential-references flags: flag provided but not defined: -badflag"))
				})
			})

			Context("when the product-name flag is not provided", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("error: product-name is missing. Please see usage for more information."))
				})
			})

			Context("when the deployed product cannot be found", func() {
				BeforeEach(func() {
					dpLister.DeployedProductsReturns([]api.DeployedProductOutput{}, nil)
				})

				It("returns an error", func() {
					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references")))
				})
			})

			Context("when there are no credential references to list", func() {
				It("prints a helpful message instead of a table", func() {
					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)

					crService.ListReturns(api.CredentialReferencesOutput{}, nil)

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(tableWriter.SetHeaderCallCount()).To(Equal(0))
					Expect(tableWriter.AppendCallCount()).To(Equal(0))
					Expect(tableWriter.RenderCallCount()).To(Equal(0))

					Expect(logger.PrintfArgsForCall(0)).To(Equal("no credential references found"))
				})
			})

			Context("when the credential references cannot be fetched", func() {
				It("returns an error", func() {
					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)

					crService.ListReturns(api.CredentialReferencesOutput{}, errors.New("could not fetch credential references"))

					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references: could not fetch credential references")))

					Expect(tableWriter.SetHeaderCallCount()).To(Equal(0))
					Expect(tableWriter.AppendCallCount()).To(Equal(0))
					Expect(tableWriter.RenderCallCount()).To(Equal(0))
				})
			})

			Context("when the deployed products cannot be fetched", func() {
				It("returns an error", func() {
					dpLister.DeployedProductsReturns(
						[]api.DeployedProductOutput{},
						errors.New("could not fetch deployed products"))

					command := commands.NewCredentialReferences(crService, dpLister, tableWriter, logger)
					err := command.Execute([]string{
						"--product-name", "some-product",
					})
					Expect(err).To(MatchError(ContainSubstring("failed to list credential references: could not fetch deployed products")))

					Expect(tableWriter.SetHeaderCallCount()).To(Equal(0))
					Expect(tableWriter.AppendCallCount()).To(Equal(0))
					Expect(tableWriter.RenderCallCount()).To(Equal(0))
				})
			})

		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewCredentialReferences(nil, nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command lists credential references for deployed products.",
				ShortDescription: "list credential references for a deployed product",
				Flags:            command.Options,
			}))
		})
	})
})
