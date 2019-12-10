package commands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"log"
	"regexp"
	"strings"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ProductDiff", func() {

	var (
		logBuffer *gbytes.Buffer
		logger    *log.Logger
		service   *fakes.ProductDiffService
		err       error
	)

	BeforeEach(func() {
		service = &fakes.ProductDiffService{}
		logBuffer = gbytes.NewBuffer()
		logger = log.New(logBuffer, "", 0)
	})

	When("a product is provided", func() {
		When("there are both manifest and runtime config differences", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
							},
							{
								Name:   "example-same-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-also-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: another-runtime-config\n   jobs:\n   - name: another-job\n     properties:\n+      timeout: 110\n-      timeout: 31",
							},
						},
					}, nil)
			})

			It("succeeds", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
			})

			It("prints the product manifest diff", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.ProductDiffArgsForCall(0)).To(Equal("example-product"))
				Expect(logBuffer).To(gbytes.Say(regexp.QuoteMeta("properties:\n+  host: example.com\n-  host: localhost\n")))
			})

			It("prints the runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				expectedOutput := `## Product Manifest

 properties:
+  host: example.com
-  host: localhost

## Runtime Configs

### example-different-runtime-config

 addons:
 - name: a-runtime-config
   jobs:
   - name: a-job
     properties:
+      timeout: 100
-      timeout: 30

### example-also-different-runtime-config

 addons:
 - name: another-runtime-config
   jobs:
   - name: another-job
     properties:
+      timeout: 110
-      timeout: 31
`
				decolorize, err := regexp.Compile(`\x1b[[0-9;]*m`)
				Expect(err).NotTo(HaveOccurred())

				contents := strings.Split(string(logBuffer.Contents()), "\n")
				for i := range contents {
					contents[i] = decolorize.ReplaceAllLiteralString(contents[i], "")
				}
				Expect(strings.Join(contents, "\n")).To(ContainSubstring(expectedOutput))

			})
		})

		When("there are product manifest changes only", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-same-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-also-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
						},
					}, nil)
			})
			It("says there are no runtime config differences and prints manifest diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("host: example.com"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
			})
		})

		When("there are runtime config changes only", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "same",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
							},
						},
					}, nil)
			})

			It("says there are no product manifest differences and prints runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("timeout: 30"))

			})
		})

		When("there are neither manifest or runtime config changes", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "same",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
						},
					}, nil)
			})
			It("says there are no manifest differences and no runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
			})
		})

		When("the product is an addon tile with no manifest", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "does_not_exist",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 90",
							},
						},
					}, nil)
			})

			It("says there is no manifest for the product and prints runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no manifest for this product"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("timeout: 90"))

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

	When("no product is provided", func() {
		It("returns a validation error", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`could not parse product-diff flags: missing required flag "--product"`))
		})
	})
})
