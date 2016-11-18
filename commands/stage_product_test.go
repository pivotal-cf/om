package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StageProduct", func() {
	var (
		stagedProductsService    *fakes.ProductStager
		availableProductsService *fakes.AvailableProductChecker
		logger                   *fakes.Logger
	)

	BeforeEach(func() {
		stagedProductsService = &fakes.ProductStager{}
		availableProductsService = &fakes.AvailableProductChecker{}
		logger = &fakes.Logger{}
	})

	It("stages a product", func() {
		availableProductsService.CheckProductAvailabilityReturns(true, nil)

		command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)

		err := command.Execute([]string{
			"--product-name", "some-product",
			"--product-version", "some-version",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(stagedProductsService.StageCallCount()).To(Equal(1))
		Expect(stagedProductsService.StageArgsForCall(0)).To(Equal(api.StageProductInput{
			ProductName:    "some-product",
			ProductVersion: "some-version",
		}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("staging some-product some-version"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished staging"))
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse stage-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product-name flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)
				err := command.Execute([]string{"--product-version", "1.0"})
				Expect(err).To(MatchError("error: product-name is missing. Please see usage for more information."))
			})
		})

		Context("when the product-version flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)
				err := command.Execute([]string{"--product-name", "some-product"})
				Expect(err).To(MatchError("error: product-version is missing. Please see usage for more information."))
			})
		})

		Context("when the product is not available", func() {
			BeforeEach(func() {
				availableProductsService.CheckProductAvailabilityReturns(false, nil)
			})

			It("returns an error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)

				err := command.Execute([]string{
					"--product-name", "some-product",
					"--product-version", "some-version",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to stage product: cannot find product")))
			})
		})

		Context("when the product availability cannot be determined", func() {
			BeforeEach(func() {
				availableProductsService.CheckProductAvailabilityReturns(false, errors.New("failed to check availability"))
			})

			It("returns an error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)

				err := command.Execute([]string{
					"--product-name", "some-product",
					"--product-version", "some-version",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to stage product: cannot check availability")))
			})
		})

		Context("when the product cannot be staged", func() {
			It("returns and error", func() {
				command := commands.NewStageProduct(stagedProductsService, availableProductsService, logger)
				availableProductsService.CheckProductAvailabilityReturns(true, nil)
				stagedProductsService.StageReturns(errors.New("some product error"))

				err := command.Execute([]string{"--product-name", "some-product", "--product-version", "some-version"})
				Expect(err).To(MatchError("failed to stage product: some product error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStageProduct(nil, nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command attempts to stage a product in the Ops Manager",
				ShortDescription: "stages a given product in the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
