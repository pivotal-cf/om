package commands_test

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("BoshDiff", func() {
	var (
		logBuffer *gbytes.Buffer
		logger    *log.Logger
		service   *fakes.BoshDiffService
		err       error
	)

	BeforeEach(func() {
		service = &fakes.BoshDiffService{}
		logBuffer = gbytes.NewBuffer()
		logger = log.New(logBuffer, "", 0)
	})

	When("the --director flag is provided", func() {
		When("there is a director manifest diff", func() {
			BeforeEach(func() {
				service.DirectorDiffReturns(
					api.DirectorDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						CloudConfig: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  property: new-cloud-value\n-  property: old-cloud-value",
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
								IAASConfigurationName: "director_cpi",
								Status:                "different",
								Diff:                  " properties:\n+  property: new-cpi-value\n-  property: old-cpi-value",
							},
						},
					}, nil)
			})

			It("prints the diff with colors", func() {
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--director"})
				Expect(err).NotTo(HaveOccurred())

				bufferContents := string(logBuffer.Contents())

				Expect(logBuffer).To(gbytes.Say("## Director Manifest"))
				Expect(logBuffer).To(gbytes.Say("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  host: example.com")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  host: localhost")))

				Expect(logBuffer).To(gbytes.Say("## Director Cloud Config"))
				Expect(logBuffer).To(gbytes.Say("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  property: new-cloud-value")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  property: old-cloud-value")))

				Expect(logBuffer).To(gbytes.Say("## Director Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("### director_runtime"))
				Expect(logBuffer).To(gbytes.Say("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  property: new-value")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  property: old-value")))

				Expect(logBuffer).To(gbytes.Say("## Director CPI Configs"))
				Expect(logBuffer).To(gbytes.Say("### director_cpi"))
				Expect(logBuffer).To(gbytes.Say("properties:"))
				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  property: new-cpi-value")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  property: old-cpi-value")))
			})

			It("Errors if --check is enabled", func() {
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--director", "--check"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Differences exist between the staged and deployed versions of the requested products"))
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

				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
			})

			It("does not error if --check is passed", func() {
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product", "--check"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "example-product"})
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
				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--product-name", "err-product"})
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
			diff := commands.NewBoshDiff(service, logger)
			err = diff.Execute([]string{"--product-name", "example-product", "--product-name", "another-product"})
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

	When("specific --product and --director are not provided", func() {
		When("There are changes to the director and product", func() {
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
				diff := commands.NewBoshDiff(service, logger)
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
		})

		When("there are changes to the director but not the product", func() {
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
							Diff:   " instance_groups:\n - name: bosh\n   properties:\n     hm:\n+      tsdb_enabled: \"<redacted>\"\n+      tsdb:\n+        address: \"<redacted>\"\n+        port: \"<redacted>\"",
						},
						CloudConfig: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  property: new-cloud-value\n-  property: old-cloud-value",
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
								IAASConfigurationName: "director_cpi",
								Status:                "different",
								Diff:                  " properties:\n+  property: new-cpi-value\n-  property: old-cpi-value",
							},
						},
					}, nil)
				service.ProductDiffReturns(api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "same",
						Diff:   "",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
				}, nil)
			})

			It("--check prints runtime and manifest configs for the director and the product and exits with 1", func() {
				//disable color for just this test;
				//we don't want to try to assemble this whole example with color
				color.NoColor = true
				defer func() { color.NoColor = false }()

				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{"--check"})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Differences exist between the staged and deployed versions of the requested products"))
				expectedOutput := `## Director Manifest

 instance_groups:
 - name: bosh
   properties:
     hm:
+      tsdb_enabled: "<redacted>"
+      tsdb:
+        address: "<redacted>"
+        port: "<redacted>"

## Director Cloud Config

 properties:
+  property: new-cloud-value
-  property: old-cloud-value

## Director Runtime Configs

### director_runtime

 properties:
+  property: new-value
-  property: old-value

## Director CPI Configs

### director_cpi

 properties:
+  property: new-cpi-value
-  property: old-cpi-value
`
				Expect(string(logBuffer.Contents())).To(ContainSubstring(expectedOutput))
			})
		})

		When("there is an error from the DirectorDiff method", func() {
			It("returns that error", func() {
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
				service.DirectorDiffReturns(api.DirectorDiff{}, fmt.Errorf("insufficient cooks"))

				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{""})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("could not discover the director diff: insufficient cooks"))
			})
		})

		When("there is an error from the ListStagedProducts method", func() {
			It("returns that error", func() {
				service.ListStagedProductsReturns(
					api.StagedProductsOutput{}, fmt.Errorf("insufficient cooks"))

				diff := commands.NewBoshDiff(service, logger)
				err = diff.Execute([]string{""})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("could not discover staged products to diff: insufficient cooks"))
			})
		})
	})
})
