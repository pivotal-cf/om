package commands_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigTemplate", func() {
	var (
		command *commands.ConfigTemplate
	)

	createOutputDirectory := func() string {
		tempDir, err := os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())

		return tempDir
	}

	Describe("Execute", func() {
		BeforeEach(func() {
			command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
				f := &fakes.MetadataProvider{}
				f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
				return f
			})
		})

		Describe("upserting an entry in the output directory with template files", func() {
			When("the output directory does not exist", func() {
				It("returns an error indicating the path does not exist", func() {
					args := []string{
						"--output-directory", "/not/real/directory",
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "c",
						"--product-version", "d",
					}

					err := executeCommand(command, args)
					Expect(err).To(MatchError("output-directory does not exist: /not/real/directory"))
				})

			})

			When("the output directory already exists without the product's directory", func() {
				var (
					tempDir string
					args    []string
				)

				When("--exclude-version is not set", func() {
					BeforeEach(func() {
						tempDir = createOutputDirectory()

						args = []string{
							"--output-directory", tempDir,
							"--pivnet-api-token", "b",
							"--pivnet-product-slug", "c",
							"--product-version", "d",
						}
					})

					It("creates nested subdirectories named by product slug and version", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product")
						versionDir := filepath.Join(productDir, "1.1.1")

						Expect(productDir).To(BeADirectory())
						Expect(versionDir).To(BeADirectory())
					})

					It("creates the various generated sub directories within the product directory", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						featuresDir := filepath.Join(tempDir, "example-product", "1.1.1", "features")
						Expect(featuresDir).To(BeADirectory())

						networkDir := filepath.Join(tempDir, "example-product", "1.1.1", "network")
						Expect(networkDir).To(BeADirectory())

						optionalDir := filepath.Join(tempDir, "example-product", "1.1.1", "optional")
						Expect(optionalDir).To(BeADirectory())

						resourceDir := filepath.Join(tempDir, "example-product", "1.1.1", "resource")
						Expect(resourceDir).To(BeADirectory())
					})

					It("creates the correct files", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product", "1.1.1")

						Expect(filepath.Join(productDir, "errand-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "product.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "default-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "required-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "resource-vars.yml")).To(BeAnExistingFile())
					})
				})

				When("--exclude-version is set", func() {
					BeforeEach(func() {
						tempDir = createOutputDirectory()

						args = []string{
							"--output-directory", tempDir,
							"--pivnet-api-token", "b",
							"--pivnet-product-slug", "c",
							"--product-version", "d",
							"--exclude-version",
						}
					})

					It("creates nested subdirectories named by product slug", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product")
						versionDir := filepath.Join(productDir, "1.1.1")

						Expect(productDir).To(BeADirectory())
						Expect(versionDir).ToNot(BeADirectory())
					})

					It("creates the various generated sub directories within the product directory", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						featuresDir := filepath.Join(tempDir, "example-product", "features")
						Expect(featuresDir).To(BeADirectory())

						networkDir := filepath.Join(tempDir, "example-product", "network")
						Expect(networkDir).To(BeADirectory())

						optionalDir := filepath.Join(tempDir, "example-product", "optional")
						Expect(optionalDir).To(BeADirectory())

						resourceDir := filepath.Join(tempDir, "example-product", "resource")
						Expect(resourceDir).To(BeADirectory())
					})

					It("creates the correct files", func() {
						err := executeCommand(command, args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product")

						Expect(filepath.Join(productDir, "errand-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "product.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "default-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "required-vars.yml")).To(BeAnExistingFile())
						Expect(filepath.Join(productDir, "resource-vars.yml")).To(BeAnExistingFile())
					})
				})
			})
		})

		When("the product has a collection", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`---
name: example-product
product_version: "1.1.1"
form_types:
- property_inputs:
  - reference: .properties.some_property
property_blueprints:
- type: collection
  optional: true
  name: some_property
  configurable: true
  property_blueprints:
  - name: name
    type: string
`), nil)
					return f
				})
			})

			It("outputs 10 ops-files by default", func() {
				tempDir := createOutputDirectory()

				err := executeCommand(command, []string{
					"--output-directory", tempDir,
					"--pivnet-api-token", "b",
					"--pivnet-product-slug", "c",
					"--product-version", "d",
				})
				Expect(err).ToNot(HaveOccurred())

				matches, err := filepath.Glob(filepath.Join(tempDir, "example-product", "1.1.1", "optional", "*.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(matches).To(HaveLen(10))
			})

			It("outputs 1 ops-file per collection when specifying --size-of-collections", func() {
				tempDir := createOutputDirectory()

				err := executeCommand(command, []string{
					"--output-directory", tempDir,
					"--pivnet-api-token", "b",
					"--pivnet-product-slug", "c",
					"--product-version", "d",
					"--size-of-collections", "10",
				})
				Expect(err).ToNot(HaveOccurred())

				expectedContents := `- type: replace
  path: /product-properties/.properties.some_property?
  value:
    value:
    - name: ((some_property_0_name))
    - name: ((some_property_1_name))
    - name: ((some_property_2_name))
    - name: ((some_property_3_name))
    - name: ((some_property_4_name))
    - name: ((some_property_5_name))
    - name: ((some_property_6_name))
    - name: ((some_property_7_name))
    - name: ((some_property_8_name))
    - name: ((some_property_9_name))
`

				matches, err := filepath.Glob(filepath.Join(tempDir, "example-product", "1.1.1", "optional", "*.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(matches).To(HaveLen(1))
				contents, err := os.ReadFile(matches[0])
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(MatchYAML(expectedContents))
			})
		})
	})

	Describe("flag handling", func() {
		When("pivnet and product path args are provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			It("returns an error", func() {
				err := executeCommand(command, []string{
					"--output-directory", createOutputDirectory(),
					"--pivnet-api-token", "b",
					"--product-path", "c",
				})
				Expect(err).To(MatchError(ContainSubstring("please provide either pivnet flags OR product-path")))
			})
		})

		When("the cli args arg not provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			DescribeTable("returns an error", func(required, message string) {
				args := []string{
					"--output-directory", createOutputDirectory(),
					"--pivnet-api-token", "b",
					"--pivnet-product-slug", "c",
					"--product-version", "d",
				}
				for i, value := range args {
					if value == required {
						args = append(args[0:i], args[i+2:]...)
						break
					}
				}

				err := executeCommand(command, args)
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
				Entry("with pivnet-api-token", "--pivnet-api-token", "please provide either pivnet flags OR product-path"),
				Entry("with pivnet-product-slug", "--pivnet-product-slug", "please provide either pivnet flags OR product-path"),
				Entry("with product-version", "--product-version", "please provide either pivnet flags OR product-path"),
			)
		})

		Describe("metadata extraction and parsing failures", func() {
			When("the metadata cannot be extracted", func() {
				BeforeEach(func() {
					command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
						f := &fakes.MetadataProvider{}
						f.MetadataBytesReturns(nil, errors.New("cannot get metadata"))
						return f
					})
				})

				It("returns an error", func() {
					tempDir := createOutputDirectory()

					args := []string{
						"--output-directory", tempDir,
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "example-product",
						"--product-version", "1.1.1",
					}

					err := executeCommand(command, args)
					Expect(err).To(MatchError("error getting metadata for example-product at version 1.1.1: cannot get metadata"))
				})
			})
			When("The returned metadata's version is an empty string", func() {
				BeforeEach(func() {
					command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
						f := &fakes.MetadataProvider{}
						f.MetadataBytesReturns([]byte(`{name: example-product, product_version: ""}`), nil)
						return f
					})
				})
				It("errors", func() {
					tempDir := createOutputDirectory()

					args := []string{
						"--output-directory", tempDir,
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "example-product",
						"--product-version", "1.1.1",
					}

					err := executeCommand(command, args)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
