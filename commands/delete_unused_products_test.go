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

var _ = Describe("DeleteProduct", func() {
	var (
		command     *commands.DeleteUnusedProducts
		fakeService *fakes.DeleteUnusedProductsService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.DeleteUnusedProductsService{}
		logger = &fakes.Logger{}
		command = commands.NewDeleteUnusedProducts(fakeService, logger)
	})

	Describe("Execute", func() {
		It("deletes all the product", func() {
			err := executeCommand(command,[]string{})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.DeleteAvailableProductsCallCount()).To(Equal(1))

			input := fakeService.DeleteAvailableProductsArgsForCall(0)
			Expect(input).To(Equal(api.DeleteAvailableProductsInput{
				ShouldDeleteAllProducts: true,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("trashing unused products"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("done"))
		})
	})

	When("an error occurs", func() {
		When("deleting all products fails", func() {
			It("returns an error", func() {
				fakeService.DeleteAvailableProductsReturns(errors.New("something bad happened"))

				err := executeCommand(command,[]string{"-p", "nah", "-v", "nope"})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
