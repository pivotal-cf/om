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

var _ = Describe("DeleteProduct", func() {
	var (
		command commands.DeleteUnusedProducts
		deleter *fakes.ProductDeleter
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		deleter = &fakes.ProductDeleter{}
		logger = &fakes.Logger{}
		command = commands.NewDeleteUnusedProducts(deleter, logger)
	})

	Describe("Execute", func() {
		It("deletes all the product", func() {
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(deleter.DeleteCallCount()).To(Equal(1))

			input, deleteAll := deleter.DeleteArgsForCall(0)
			Expect(input).To(Equal(api.AvailableProductsInput{}))
			Expect(deleteAll).To(BeTrue())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("trashing unused products"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("done"))
		})
	})

	Context("when an error occurs", func() {
		Context("when deleting all products fails", func() {
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
			Expect(usage).To(Equal(jhanda.Usage{
				Description:      "This command deletes unused products in the targeted Ops Manager",
				ShortDescription: "deletes unused products on the Ops Manager targeted",
			}))
		})
	})
})
