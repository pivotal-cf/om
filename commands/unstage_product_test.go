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

var _ = Describe("UnstageProduct", func() {
	var (
		fakeService *fakes.UnstageProductService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.UnstageProductService{}
		logger = &fakes.Logger{}
	})

	It("unstages a product", func() {
		command := commands.NewUnstageProduct(fakeService, logger)

		err := command.Execute([]string{
			"--product-name", "some-product",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(fakeService.DeleteStagedProductCallCount()).To(Equal(1))
		Expect(fakeService.DeleteStagedProductArgsForCall(0)).To(Equal(
			api.UnstageProductInput{
				ProductName: "some-product",
			}))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("unstaging some-product"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished unstaging"))
	})

	Context("failure cases", func() {
		When("an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse unstage-product flags: flag provided but not defined: -badflag"))
			})
		})

		When("the product-name flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse unstage-product flags: missing required flag \"--product-name\""))
			})
		})

		When("the product cannot be unstaged", func() {
			It("returns an error", func() {
				command := commands.NewUnstageProduct(fakeService, logger)
				fakeService.DeleteStagedProductReturns(errors.New("some product error"))

				err := command.Execute([]string{"--product-name", "some-product"})
				Expect(err).To(MatchError("failed to unstage product: some product error"))
			})
		})

	})
})
