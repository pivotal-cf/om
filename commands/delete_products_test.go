package commands_test

import (
	"errors"

	jhandacommands "github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteProduct", func() {
	var (
		command commands.DeleteProduct
		deleter *fakes.ProductDeleter
	)

	BeforeEach(func() {
		deleter = &fakes.ProductDeleter{}
		command = commands.NewDeleteProduct(deleter)
	})

	Describe("Execute", func() {
		It("deletes the specific product", func() {
			err := command.Execute([]string{"-p", "some-product-name", "-v", "1.2.3-build.4"})
			Expect(err).NotTo(HaveOccurred())

			Expect(deleter.DeleteCallCount()).To(Equal(1))

			input, deleteAll := deleter.DeleteArgsForCall(0)
			Expect(input).To(Equal(api.AvailableProductsInput{
				ProductName:    "some-product-name",
				ProductVersion: "1.2.3-build.4",
			}))
			Expect(deleteAll).To(BeFalse())
		})
	})

	Context("when an error occurs", func() {
		Context("when deleting a product fails", func() {
			It("returns an error", func() {
				deleter.DeleteReturns(errors.New("something bad happened"))

				err := command.Execute([]string{"-p", "nah", "-v", "nope"})
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage", func() {
			usage := command.Usage()
			Expect(usage).To(Equal(jhandacommands.Usage{
				Description:      "This command deletes the named product from the targeted Ops Manager",
				ShortDescription: "deletes a product from the Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
