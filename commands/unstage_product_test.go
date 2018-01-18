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
)

var _ = Describe("UnstageProduct", func() {
	var (
		stagedProductsService *fakes.ProductUnstager
		logger                *fakes.Logger
	)

	BeforeEach(func() {
		stagedProductsService = &fakes.ProductUnstager{}
		logger = &fakes.Logger{}
	})

	It("unstages a product", func() {
		command := commands.NewUnstageProduct(stagedProductsService, logger)

		err := command.Execute([]string{
			"--product-name", "some-product",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(stagedProductsService.UnstageCallCount()).To(Equal(1))
		Expect(stagedProductsService.UnstageArgsForCall(0)).To(Equal(
			api.UnstageProductInput{
				ProductName: "some-product",
			}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("unstaging some-product"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished unstaging"))
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(stagedProductsService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse unstage-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product-name flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(stagedProductsService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse unstage-product flags: missing required flag \"--product-name\""))
			})
		})

		Context("when the product cannot be unstaged", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(stagedProductsService, logger)
				stagedProductsService.UnstageReturns(errors.New("some product error"))

				err := command.Execute([]string{"--product-name", "some-product"})
				Expect(err).To(MatchError("failed to unstage product: some product error"))
			})
		})

	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewUnstageProduct(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command attempts to unstage a product from the Ops Manager",
				ShortDescription: "unstages a given product from the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
