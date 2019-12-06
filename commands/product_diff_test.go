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

	When("a valid product is provided", func() {
		BeforeEach(func() {
			service.ProductDiffReturns(
				api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff:   "properties:\n+  host: example.com\n-  host: localhost",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{
						{
							Name:   "example-different-runtime-config",
							Status: "different",
							Diff:   "addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
						},
						{
							Name:   "example-same-runtime-config",
							Status: "same",
							Diff:   "",
						},
						{
							Name:   "example-also-different-runtime-config",
							Status: "different",
							Diff:   "addons:\n - name: another-runtime-config\n   jobs:\n   - name: another-job\n     properties:\n+      timeout: 110\n-      timeout: 31",
						},
					},
				}, nil)
		})

		It("succeeds", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("prints the product manifest diff", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "example-product"})
			Expect(err).NotTo(HaveOccurred())
			Expect(service.ProductDiffArgsForCall(0)).To(Equal("example-product"))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: different")))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("properties:\n+  host: example.com\n-  host: localhost\n")))
		})

		PIt("prints the runtime config diffs", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "example-product"})
			Expect(err).NotTo(HaveOccurred())
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: different")))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: different")))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: same")))
			Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("Status: different")))

		})
	})
	When("there is an error from the diff service", func() {
		It("returns that error", func() {
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
