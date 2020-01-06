package commands_test

import (
	"fmt"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"log"
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

	PWhen("the --director flag is provided", func() {
		When("there is a director manifest diff", func() {
			BeforeEach(func() {
				service.DirectorDiffReturns(
					api.DirectorDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						CloudConfig: api.ManifestDiff{
							Status: "same",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "director_runtime",
								Status: "different",
								Diff:   " properties:\n+  property: new-value\n-  property: old-value",
							},
						},
						CPIConfigs: []api.CPIConfigsDiff{
							{
								IAASConfigurationName: "default",
								Status:                "different",
								Diff:                  ` properties: datacenters: - name: "<redacted>" clusters: + - canada: {}`,
							},
						},
					}, nil)
			})

			It("prints the diff with colors", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--director"})
				Expect(err).NotTo(HaveOccurred())

				bufferContents := string(logBuffer.Contents())

				Expect(bufferContents).To(ContainSubstring("## Director Manifest"))
				Expect(bufferContents).To(ContainSubstring("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  host: example.com")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  host: localhost")))

				Expect(bufferContents).To(ContainSubstring("## Director Runtime Configs"))
				Expect(bufferContents).To(ContainSubstring("### director_runtime"))
				Expect(bufferContents).To(ContainSubstring("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  property: new-value")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  property: old-value")))

			})
		})
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

			It("prints both", func() {
				//disable color for just this test;
				//we don't want to try to assemble this whole example with color
				color.NoColor = true
				defer func() { color.NoColor = false }()

				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				expectedOutput := `## Product Manifest for example-product

 properties:
+  host: example.com
-  host: localhost

## Runtime Configs for example-product

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
				Expect(string(logBuffer.Contents())).To(ContainSubstring(expectedOutput))
			})

			It("has colors on the diff", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())

				bufferContents := string(logBuffer.Contents())

				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  host: example.com")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  host: localhost")))

				Expect(bufferContents).To(ContainSubstring(color.GreenString("+      timeout: 110")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-      timeout: 31")))
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

		When("the product is staged for initial installation", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "to_be_installed",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{},
					}, nil)
			})

			It("says the product will be installed for the first time", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("This product is not yet deployed, so the product and runtime diffs are not available."))
				Expect(logBuffer).NotTo(gbytes.Say("## Runtime Configs"))
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
				Expect(err.Error()).To(Equal("too many cooks"))
				Expect(service.ProductDiffArgsForCall(0)).To(Equal("err-product"))
			})
		})
	})

	When("providing multiple products", func() {
		BeforeEach(func() {
			service.ProductDiffReturnsOnCall(0,
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

			service.ProductDiffReturnsOnCall(1,
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

		It("prints both product statuses", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "example-product", "--product", "another-product"})
			Expect(err).NotTo(HaveOccurred())

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for example-product"))
			Expect(logBuffer).To(gbytes.Say("properties:"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for example-product"))
			Expect(logBuffer).To(gbytes.Say("example-different-runtime-config"))

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for another-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for another-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))
		})
	})

	When("no product and director is provided", func() {
		BeforeEach(func() {
			service.ListStagedProductsReturns(api.StagedProductsOutput{Products: []api.StagedProduct{{
				GUID: "p-bosh-guid",
				Type: "p-bosh",
			}, {
				GUID: "example-product-guid",
				Type: "example-product",
			}, {
				GUID: "another-product-guid",
				Type: "another-product",
			}}}, nil)

			service.DirectorDiffReturns(
				api.DirectorDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff:   " properties:\n+  host: example.com\n-  host: localhost",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
					CPIConfigs:     []api.CPIConfigsDiff{},
				}, nil)

			service.ProductDiffReturnsOnCall(0,
				api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff:   " properties:\n+  host: example.com\n-  host: localhost",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
				}, nil)

			service.ProductDiffReturnsOnCall(1,
				api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff:   " properties:\n+  host: example.net\n-  host: localhost",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
				}, nil)
		})

		It("lists all staged products (alphabetically by name) as well as the director", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(string(logBuffer.Contents())).NotTo(ContainSubstring("p-bosh"))

			Expect(logBuffer).To(gbytes.Say("## Director Manifest"))
			Expect(logBuffer).To(gbytes.Say("properties:"))

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for another-product"))
			Expect(logBuffer).To(gbytes.Say("properties:"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for another-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for example-product"))
			Expect(logBuffer).To(gbytes.Say("properties:"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for example-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))
		})

		When("there is an error from the ListStagedProducts method", func() {
			It("returns that error", func() {
				// setup
				service.ListStagedProductsReturns(
					api.StagedProductsOutput{}, fmt.Errorf("insufficient cooks"))

				// execute
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{""})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("could not discover staged products to diff: insufficient cooks"))
			})
		})
	})
})
