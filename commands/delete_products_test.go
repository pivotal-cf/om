package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteProduct", func() {
	var (
		command     *commands.DeleteProduct
		fakeService *fakes.DeleteProductService
	)

	BeforeEach(func() {
		fakeService = &fakes.DeleteProductService{}
		command = commands.NewDeleteProduct(fakeService)
	})

	Describe("Execute", func() {
		It("deletes the specific product", func() {
			err := executeCommand(command, []string{"-p", "some-product-name", "-v", "1.2.3-build.4"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.DeleteAvailableProductsCallCount()).To(Equal(1))

			input := fakeService.DeleteAvailableProductsArgsForCall(0)
			Expect(input).To(Equal(api.DeleteAvailableProductsInput{
				ProductName:             "some-product-name",
				ProductVersion:          "1.2.3-build.4",
				ShouldDeleteAllProducts: false,
			}))
		})

		When("deleting a product fails", func() {
			It("returns an error", func() {
				fakeService.DeleteAvailableProductsReturns(errors.New("something bad happened"))

				err := executeCommand(command, []string{"-p", "nah", "-v", "nope"})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
