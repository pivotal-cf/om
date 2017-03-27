package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errands", func() {
	var (
		tableWriter          *fakes.TableWriter
		stagedProductsFinder *fakes.StagedProductsFinder
		errandsService       *fakes.ErrandsService
		command              commands.Errands
	)

	BeforeEach(func() {
		tableWriter = &fakes.TableWriter{}
		stagedProductsFinder = &fakes.StagedProductsFinder{}
		errandsService = &fakes.ErrandsService{}
		command = commands.NewErrands(tableWriter, errandsService, stagedProductsFinder)
	})

	Describe("Execute", func() {
		It("lists the available products", func() {
			errandsService.ListReturns(api.ErrandsListOutput{
				Errands: []api.Errand{
					{Name: "first-errand", PostDeploy: "true"},
					{Name: "second-errand", PostDeploy: "false"},
					{Name: "third-errand", PreDelete: "true"},
				},
			}, nil)

			stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{
				Product: api.StagedProduct{
					Type: "some-product-name",
					GUID: "some-product-id",
				},
			}, nil)

			err := command.Execute([]string{"--product-name", "some-product-name"})
			Expect(err).NotTo(HaveOccurred())

			Expect(stagedProductsFinder.FindCallCount()).To(Equal(1))
			Expect(stagedProductsFinder.FindArgsForCall(0)).To(Equal("some-product-name"))

			Expect(errandsService.ListCallCount()).To(Equal(1))
			Expect(errandsService.ListArgsForCall(0)).To(Equal("some-product-id"))

			Expect(tableWriter.SetHeaderCallCount()).To(Equal(1))
			Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Post Deploy Enabled", "Pre Delete Enabled"}))

			Expect(tableWriter.AppendCallCount()).To(Equal(3))
			Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"first-errand", "true", ""}))
			Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{"second-errand", "false", ""}))
			Expect(tableWriter.AppendArgsForCall(2)).To(Equal([]string{"third-errand", "", "true"}))

			Expect(tableWriter.RenderCallCount()).To(Equal(1))
		})

		Context("failure cases", func() {
			Context("when an unknown flag is passed", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--unknown-flag"})
					Expect(err).To(MatchError("could not parse errands flags: flag provided but not defined: -unknown-flag"))
				})
			})

			Context("when the staged products finder fails", func() {
				It("returns an error", func() {
					stagedProductsFinder.FindReturns(api.StagedProductsFindOutput{}, errors.New("there was an error"))
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("failed to find staged product \"some-product\": there was an error"))
				})
			})

			Context("when the errands service fails", func() {
				It("returns an error", func() {
					errandsService.ListReturns(api.ErrandsListOutput{}, errors.New("there was an error"))
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("failed to list errands: there was an error"))
				})
			})

			Context("when the product name is missing", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("error: product-name is missing. Please see usage for more information."))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewErrands(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command lists all errands for a product.",
				ShortDescription: "list errands for a product",
			}))
		})
	})
})
