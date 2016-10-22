package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	commonfakes "github.com/pivotal-cf/om/common/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StageProduct", func() {
	var (
		productService *fakes.ProductService
		logger         *commonfakes.OtherLogger
	)

	BeforeEach(func() {
		productService = &fakes.ProductService{}
		logger = &commonfakes.OtherLogger{}
	})

	It("stages a product", func() {
		command := commands.NewStageProduct(productService, logger)

		err := command.Execute([]string{
			"--product-name", "some-product",
			"--product-version", "some-version",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(productService.StageCallCount()).To(Equal(1))
		Expect(productService.StageArgsForCall(0)).To(Equal(api.StageProductInput{
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
				command := commands.NewStageProduct(productService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse stage-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product-name flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(productService, logger)
				err := command.Execute([]string{"--product-version", "1.0"})
				Expect(err).To(MatchError("error: product-name is missing. Please see usage for more information."))
			})
		})

		Context("when the product-version flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewStageProduct(productService, logger)
				err := command.Execute([]string{"--product-name", "some-product"})
				Expect(err).To(MatchError("error: product-version is missing. Please see usage for more information."))
			})
		})

		Context("when the product cannot be staged", func() {
			It("returns and error", func() {
				command := commands.NewStageProduct(productService, logger)
				productService.StageReturns(errors.New("some product error"))

				err := command.Execute([]string{"--product-name", "some-product", "--product-version", "some-version"})
				Expect(err).To(MatchError("failed to stage product: some product error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStageProduct(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command attempts to stage a product in the Ops Manager",
				ShortDescription: "stages a given product in the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
