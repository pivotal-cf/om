package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ProductDiff", func() {

	var (
		logger *fakes.Logger
		service *fakes.ProductDiffService
		err error
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		service = &fakes.ProductDiffService{}
	})

	Context("a valid product is provided", func(){

		It("succeeds", func(){
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
		})

		PIt("prints the product manifest diff", func(){

		})

		PIt("prints the runtime config diffs", func(){

		})
	})

})
