package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AvailableProducts", func() {
	var (
		apService   *fakes.AvailableProductsService
		tableWriter *fakes.TableWriter
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		apService = &fakes.AvailableProductsService{}
		tableWriter = &fakes.TableWriter{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("lists the available products", func() {
			command := commands.NewAvailableProducts(apService, tableWriter, logger)

			apService.ListReturns(api.AvailableProductsOutput{
				ProductsList: []api.ProductInfo{
					api.ProductInfo{
						Name:    "first-product",
						Version: "1.2.3",
					},
					api.ProductInfo{
						Name:    "second-product",
						Version: "4.5.6",
					},
				},
			}, nil)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tableWriter.SetHeaderArgsForCall(0)).To(Equal([]string{"Name", "Version"}))

			Expect(tableWriter.AppendCallCount()).To(Equal(2))
			Expect(tableWriter.AppendArgsForCall(0)).To(Equal([]string{"first-product", "1.2.3"}))
			Expect(tableWriter.AppendArgsForCall(1)).To(Equal([]string{"second-product", "4.5.6"}))

			Expect(tableWriter.RenderCallCount()).To(Equal(1))
		})

		Context("when there are no products to list", func() {
			It("prints a helpful message instead of a table", func() {
				command := commands.NewAvailableProducts(apService, tableWriter, logger)

				apService.ListReturns(api.AvailableProductsOutput{}, nil)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(tableWriter.SetHeaderCallCount()).To(Equal(0))
				Expect(tableWriter.AppendCallCount()).To(Equal(0))
				Expect(tableWriter.RenderCallCount()).To(Equal(0))

				Expect(logger.PrintfArgsForCall(0)).To(Equal("no available products found"))
			})
		})

		Context("error cases", func() {
			It("returns the error", func() {
				command := commands.NewAvailableProducts(apService, tableWriter, logger)

				apService.ListReturns(api.AvailableProductsOutput{}, errors.New("blargh"))

				err := command.Execute([]string{})
				Expect(err).To(MatchError("blargh"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewAvailableProducts(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command lists all available products.",
				ShortDescription: "list available products",
			}))
		})
	})
})
