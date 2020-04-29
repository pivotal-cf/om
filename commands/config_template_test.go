package commands_test

import (
	"errors"
	"fmt"
	"io/ioutil"
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
		command     *commands.ConfigTemplate
		environFunc func() []string
	)

	createOutputDirectory := func() string {
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		return tempDir
	}

	BeforeEach(func() {
		environFunc = func() []string { return nil }
	})

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

					err := command.Execute(args)
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
						err := command.Execute(args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product")
						versionDir := filepath.Join(productDir, "1.1.1")

						Expect(productDir).To(BeADirectory())
						Expect(versionDir).To(BeADirectory())
					})

					It("creates the various generated sub directories within the product directory", func() {
						err := command.Execute(args)
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
						err := command.Execute(args)
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
						err := command.Execute(args)
						Expect(err).ToNot(HaveOccurred())

						productDir := filepath.Join(tempDir, "example-product")
						versionDir := filepath.Join(productDir, "1.1.1")

						Expect(productDir).To(BeADirectory())
						Expect(versionDir).ToNot(BeADirectory())
					})

					It("creates the various generated sub directories within the product directory", func() {
						err := command.Execute(args)
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
						err := command.Execute(args)
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

				err := command.Execute([]string{
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

			It("outputs N ops-files when specifying --size-of-collections", func() {
				tempDir := createOutputDirectory()

				err := command.Execute([]string{
					"--output-directory", tempDir,
					"--pivnet-api-token", "b",
					"--pivnet-product-slug", "c",
					"--product-version", "d",
					"--size-of-collections", "3",
				})
				Expect(err).ToNot(HaveOccurred())

				matches, err := filepath.Glob(filepath.Join(tempDir, "example-product", "1.1.1", "optional", "*.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(matches).To(HaveLen(3))
			})
		})
	})

	Describe("flag handling", func() {
		When("an unknown flag is provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			It("returns an error", func() {
				err := command.Execute([]string{"--invalid"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -invalid"))
				err = command.Execute([]string{"--unreal"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -unreal"))
			})
		})

		When("pivnet and product path args are provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			It("returns an error", func() {
				err := command.Execute([]string{
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

				err := command.Execute(args)
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
				Entry("with output-directory", "--output-directory", `missing required flag "--output-directory"`),
				Entry("with pivnet-api-token", "--pivnet-api-token", "please provide either pivnet flags OR product-path"),
				Entry("with pivnet-product-slug", "--pivnet-product-slug", "please provide either pivnet flags OR product-path"),
				Entry("with product-version", "--product-version", "please provide either pivnet flags OR product-path"),
			)
		})

		When("the --config flag is passed", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			var (
				configFile *os.File
				err        error
			)

			When("the config file contains variables", func() {
				const downloadProductConfigWithVariablesTmpl = `---
pivnet-api-token: "token"
file-glob: "*.pivotal"
pivnet-product-slug: ((product-slug))
product-version: 2.0.0
output-directory: %s
`

				BeforeEach(func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).ToNot(HaveOccurred())

					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					_, err = configFile.WriteString(fmt.Sprintf(downloadProductConfigWithVariablesTmpl, tempDir))
					Expect(err).ToNot(HaveOccurred())

					err = configFile.Close()
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					err = os.RemoveAll(configFile.Name())
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns an error if missing variables", func() {
					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring("Expected to find variables")))
				})

				Context("passed in a vars-file", func() {
					var varsFile *os.File

					BeforeEach(func() {
						varsFile, err = ioutil.TempFile("", "")
						Expect(err).ToNot(HaveOccurred())

						_, err = varsFile.WriteString(`product-slug: elastic-runtime`)
						Expect(err).ToNot(HaveOccurred())

						err = varsFile.Close()
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						err = os.RemoveAll(varsFile.Name())
						Expect(err).ToNot(HaveOccurred())
					})

					It("can interpolate variables into the configuration", func() {
						err = command.Execute([]string{
							"--config", configFile.Name(),
							"--vars-file", varsFile.Name(),
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("given vars", func() {
					It("can interpolate variables into the configuration", func() {
						err = command.Execute([]string{
							"--config", configFile.Name(),
							"--var", "product-slug=elastic-runtime",
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("passed as environment variables", func() {
					BeforeEach(func() {
						environFunc = func() []string {
							return []string{"OM_VAR_product-slug='sea-slug'"}
						}

						command = commands.NewConfigTemplateWithEnvironment(func(*commands.ConfigTemplate) commands.MetadataProvider {
							f := &fakes.MetadataProvider{}
							f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
							return f
						}, environFunc)
					})

					It("can interpolate variables into the configuration", func() {
						err = command.Execute([]string{
							"--config", configFile.Name(),
							"--vars-env", "OM_VAR",
						})
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
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

					err := command.Execute(args)
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

					err := command.Execute(args)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
