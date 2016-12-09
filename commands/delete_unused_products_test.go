package commands_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("DeleteUnusedProducts", func() {
	var (
		productsService *fakes.ProductUploader
		logger          *fakes.Logger
	)

	BeforeEach(func() {
		productsService = &fakes.ProductUploader{}
		logger = &fakes.Logger{}
	})

	It("deletes unused products", func() {
		command := commands.NewDeleteUnusedProducts(productsService, logger)

		err := command.Execute([]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(productsService.TrashCallCount()).To(Equal(1))

		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("trashing unused products"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("done"))
	})

	Context("failure cases", func() {
		Context("when the trash call returns an error", func() {
			It("returns an error", func() {
				productsService.TrashReturns(errors.New("some error"))
				command := commands.NewDeleteUnusedProducts(productsService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not delete unused products: some error"))
			})
		})
	})
})
