package commands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"log"
	"regexp"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ProductDiff", func() {

	var (
		logBuffer = gbytes.NewBuffer()
		logger    = log.New(logBuffer, "", 0)
		service   *fakes.ProductDiffService
		err       error
	)

	BeforeEach(func() {
		service = &fakes.ProductDiffService{}
	})

	When("a valid product is provided", func(){

		It("succeeds", func(){
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("prints the product manifest diff", func(){
			// setup
			service.ProductDiffReturns(
				api.ProductDiff{
					Manifest:       api.ManifestDiff{
						Status: "different",
						Diff:   "properties:\n+  host: example.com\n-  host: localhost",
					},
					RuntimeConfigs: nil,
			}, nil)

			// execute
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "example-product"})
			Expect(err).NotTo(HaveOccurred())
			Expect(service.ProductDiffArgsForCall(0)).To(Equal("example-product"))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: different")))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("properties:\n+  host: example.com\n-  host: localhost\n")))
		})

		PIt("prints the runtime config diffs", func(){

		})
	})
	When("there is an error from the diff service", func(){
		It("returns that error", func(){
			// setup
			service.ProductDiffReturns(
				api.ProductDiff{}, fmt.Errorf("too many cooks"))

			// execute
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "err-product"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("too many cooks"))
			Expect(service.ProductDiffArgsForCall(0)).To(Equal("err-product"))
		})

	})

})
