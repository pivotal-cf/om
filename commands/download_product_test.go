package commands_test

import (
	"fmt"
	"github.com/pivotal-cf/om/download_clients"
	"github.com/pivotal-cf/om/download_clients/fakes"
	"github.com/pivotal-cf/om/validator"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
	log "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
)

var _ = Describe("DownloadProduct", func() {
	var (
		command              *commands.DownloadProduct
		commandArgs          []string
		logger               *loggerfakes.FakeLogger
		fakePivnetDownloader *fakes.PivnetDownloader
		environFunc          func() []string
		tempDir              string
		err                  error
		file                 *os.File
		fileContents         = "hello world"
	)

	fakePivnetFactory := func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader {
		return fakePivnetDownloader
	}

	BeforeEach(func() {
		var err error
		file, err = ioutil.TempFile("", "[product-slug,1.0.0-beta.1]product*.pivotal")
		Expect(err).ToNot(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(fileContents)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).ToNot(HaveOccurred())

		logger = &loggerfakes.FakeLogger{}
		fakePivnetDownloader = &fakes.PivnetDownloader{}
		environFunc = func() []string { return nil }
	})

	JustBeforeEach(func() {
		command = commands.NewDownloadProduct(environFunc, logger, GinkgoWriter, fakePivnetFactory, nil)
	})

	AfterEach(func() {
		err := os.Remove(file.Name())
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when the flags are set correctly", func() {
		BeforeEach(func() {
			fakePivnetDownloader.ReleaseForVersionReturnsOnCall(0, pivnet.Release{
				ID: 12345,
			}, nil)

			fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
				{
					ID:           54321,
					AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.pivotal",
					Name:         "Example Cloud Foundry",
				},
			}, nil)

			tempDir, err = ioutil.TempDir("", "om-tests-")
			Expect(err).NotTo(HaveOccurred())

			commandArgs = []string{
				"--pivnet-api-token", "token",
				"--pivnet-file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--output-directory", tempDir,
			}
		})

		AfterEach(func() {
			err = os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("downloads a product from Pivotal Network", func() {
			err = command.Execute(commandArgs)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))
		})

		Context("when a valid product-version-regex is provided", func() {
			BeforeEach(func() {
				fakePivnetDownloader.ReleasesForProductSlugReturns([]pivnet.Release{
					{
						ID:      5,
						Version: "3.0.0",
					},
					{
						ID:      4,
						Version: "1.1.11",
					},
					{
						ID:      3,
						Version: "2.1.2",
					},
					{
						ID:      2,
						Version: "2.1.1",
					},
					{
						ID:      1,
						Version: "2.0.1",
					},
				}, nil)

				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
					{
						ID:           54321,
						AWSObjectKey: "/some-account/some-bucket/cf-2.1-build.11.pivotal",
						Name:         "Example Cloud Foundry",
					},
				}, nil)

				fakePivnetDownloader.ReleaseForVersionReturnsOnCall(0, pivnet.Release{
					ID: 4,
				}, nil)

				commandArgs = []string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version-regex", `2\..\..*`,
					"--output-directory", tempDir,
				}
			})

			It("downloads the highest version matching that regex", func() {
				err = command.Execute(commandArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakePivnetDownloader.ReleasesForProductSlugCallCount()).To(Equal(1))
				Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))
				Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
				Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(1))

				slug := fakePivnetDownloader.ReleasesForProductSlugArgsForCall(0)
				Expect(slug).To(Equal("elastic-runtime"))

				slug, version := fakePivnetDownloader.ReleaseForVersionArgsForCall(0)
				Expect(slug).To(Equal("elastic-runtime"))
				Expect(version).To(Equal("2.1.2"))

				slug, releaseID := fakePivnetDownloader.ProductFilesForReleaseArgsForCall(0)
				Expect(slug).To(Equal("elastic-runtime"))
				Expect(releaseID).To(Equal(4))

				file, slug, releaseID, productFileID, _ := fakePivnetDownloader.DownloadProductFileArgsForCall(0)
				Expect(file.Name()).To(Equal(path.Join(tempDir, "cf-2.1-build.11.pivotal")))
				Expect(slug).To(Equal("elastic-runtime"))
				Expect(releaseID).To(Equal(4))
				Expect(productFileID).To(Equal(54321))

				downloadedFilePath := path.Join(tempDir, "cf-2.1-build.11.pivotal")
				Expect(downloadedFilePath).To(BeAnExistingFile())
			})

			Context("when the releases contains non-semver-compatible version", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ReleasesForProductSlugReturns([]pivnet.Release{
						{
							ID:      3,
							Version: "2.1.2",
						},
						{
							ID:      0,
							Version: "2.0.x",
						},
					}, nil)
				})

				It("ignores the version and prints a warning", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					logStr, _ := logger.InfoArgsForCall(0)
					Expect(logStr).To(Equal("warning: could not parse semver version from: 2.0.x"))
				})
			})

			When("there are no valid versions found for given product regex", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ReleasesForProductSlugReturns([]pivnet.Release{
						{
							ID:      3,
							Version: "3.1.2",
						},
					}, nil)
				})

				It("returns an error", func() {
					err = command.Execute(commandArgs)
					Expect(err).To(MatchError("no valid versions found for product 'elastic-runtime' and product version regex '2\\..\\..*'\nexisting versions: 3.1.2"))
				})
			})
		})

		Context("when the globs returns multiple files", func() {
			BeforeEach(func() {
				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
					{
						ID:           54321,
						AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.pivotal",
						Name:         "cf-2.0-build.1.pivotal",
					},
					{
						ID:           54320,
						AWSObjectKey: "/some-account/some-bucket/srt-2.0-build.1.pivotal",
						Name:         "srt-2.0-build.1.pivotal",
					},
				}, nil)
			})

			It("returns an error", func() {
				err = command.Execute(commandArgs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`the glob '*.pivotal' matches multiple files. Write your glob to match exactly one of the following:`))
			})
		})

		Context("when the stemcell-iaas flag is set", func() {
			BeforeEach(func() {
				commandArgs = []string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
					"--stemcell-iaas", "google",
				}
				fakePivnetDownloader.ReleaseForVersionReturnsOnCall(1, pivnet.Release{
					ID: 9999,
				}, nil)

				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(1, []pivnet.ProductFile{
					{
						ID:           5678,
						AWSObjectKey: "/some-account/some-bucket/light-bosh-stemcell-97.19-google-kvm-ubuntu-xenial-go_agent.tgz",
						Name:         "Example Stemcell For GCP",
					},
				}, nil)

				fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
					{Release: pivnet.DependentRelease{
						ID:      199678,
						Version: "97.19",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199677,
						Version: "97.18",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199676,
						Version: "97.17",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199675,
						Version: "97.9",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199674,
						Version: "97",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
				}, nil)
			})

			It("grabs the latest stemcell for the product that matches the glob", func() {
				err = command.Execute(commandArgs)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakePivnetDownloader.ReleaseDependenciesCallCount()).To(Equal(1))
				Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(2))
				Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(2))
				Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(2))
				Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(2))

				str, version := fakePivnetDownloader.ReleaseForVersionArgsForCall(1)
				Expect(version).To(Equal("97.19"))
				Expect(str).To(Equal("stemcells-ubuntu-xenial"))

				fakePivnetDownloader.DownloadProductFileArgsForCall(0)

				stemcellFile, slug, releaseID, fileID, _ := fakePivnetDownloader.DownloadProductFileArgsForCall(1)
				Expect(stemcellFile.Name()).To(Equal(path.Join(tempDir, "light-bosh-stemcell-97.19-google-kvm-ubuntu-xenial-go_agent.tgz")))
				Expect(slug).To(Equal("stemcells-ubuntu-xenial"))
				Expect(releaseID).To(Equal(9999))
				Expect(fileID).To(Equal(5678))

				fileName := path.Join(tempDir, "download-file.json")
				fileContent, err := ioutil.ReadFile(fileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileName).To(BeAnExistingFile())
				downloadedFilePath := path.Join(tempDir, "cf-2.0-build.1.pivotal")
				Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`
					{
						"product_path": "%s",
						"product_slug": "elastic-runtime",
						"stemcell_path": "%s",
						"stemcell_version": "97.19"
					}`, downloadedFilePath, stemcellFile.Name())))

				fileName = path.Join(tempDir, "assign-stemcell.yml")
				fileContent, err = ioutil.ReadFile(fileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileName).To(BeAnExistingFile())
				Expect(string(fileContent)).To(MatchJSON(`
					{
						"product": "elastic-runtime",
						"stemcell": "97.19"
					}`))

			})

			Context("and the product is not a tile", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ReleaseForVersionReturnsOnCall(0, pivnet.Release{
						ID: 12345,
					}, nil)

					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.tgz",
							Name:         "Example Cloud Foundry",
						},
					}, nil)
				})

				It("exits 0 and prints a warning when the product is not a tile", func() {
					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "google",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))

					infoStr, _ := logger.InfoArgsForCall(1)
					Expect(infoStr).To(Equal("the downloaded file is not a .pivotal file. Not determining and fetching required stemcell."))
				})
			})
		})

		Context("when the file is already downloaded", func() {
			setupPivnetAPIForProduct := func(shaSum string) {
				fakePivnetDownloader.ReleaseForVersionReturnsOnCall(0, pivnet.Release{
					ID: 54321,
				}, nil)

				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
					{
						ID:           54321,
						AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.pivotal",
						SHA256:       shaSum,
						Name:         "Example Cloud Foundry",
					},
				}, nil)

				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(1, []pivnet.ProductFile{
					{
						ID:           5678,
						AWSObjectKey: "/some-account/some-bucket/light-bosh-stemcell-97.19-google-kvm-ubuntu-xenial-go_agent.tgz",
						Name:         "Example Stemcell For GCP",
					},
				}, nil)

				fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
					{Release: pivnet.DependentRelease{
						ID:      199678,
						Version: "97.19",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199677,
						Version: "97.18",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199676,
						Version: "97.17",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199675,
						Version: "97.9",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
					{Release: pivnet.DependentRelease{
						ID:      199674,
						Version: "97",
						Product: pivnet.Product{
							ID:   111,
							Slug: "stemcells-ubuntu-xenial",
							Name: "Stemcells for PCF (Ubuntu Xenial)",
						},
					}},
				}, nil)
			}

			createFilePath := func() string {
				filePath := path.Join(tempDir, "cf-2.0-build.1.pivotal")
				file, err := os.Create(filePath)
				Expect(err).NotTo(HaveOccurred())
				_, err = file.WriteString("something-not-important")
				Expect(err).NotTo(HaveOccurred())
				err = file.Close()
				Expect(err).NotTo(HaveOccurred())
				return filePath
			}

			When("a sha sum is provided by Pivnet", func() {
				BeforeEach(func() {
					filePath := createFilePath()

					validator := validator.NewSHA256Calculator()
					sum, err := validator.Checksum(filePath)
					Expect(err).NotTo(HaveOccurred())
					setupPivnetAPIForProduct(sum)
				})

				It("does not download the file again", func() {
					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(0))

					logStr, _ := logger.InfoArgsForCall(0)
					Expect(logStr).To(ContainSubstring("already exists, skip downloading"))
				})

				It("does not panic when downloading the stemcell if file already downloaded", func() {
					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--stemcell-iaas", "google",
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("no sha sum is provided", func() {
				It("does not re-download the product", func() {
					createFilePath()
					setupPivnetAPIForProduct("")

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
					Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(0))

					logStr, _ := logger.InfoArgsForCall(0)
					Expect(logStr).To(ContainSubstring("already exists, skip downloading"))
				})
			})

			When("the sha is invalid", func() {
				It("downloads it, again", func() {
					createFilePath()
					setupPivnetAPIForProduct("asdfasdfasdf")

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the --config flag is passed", func() {
			var (
				configFile *os.File
				err        error
			)

			Context("when the config file contains variables", func() {
				const downloadProductConfigWithVariablesTmpl = `---
pivnet-api-token: "token"
pivnet-file-glob: "*.pivotal"
pivnet-product-slug: ((product-slug))
product-version: 2.0.0
output-directory: %s
`

				BeforeEach(func() {
					configFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					_, err = configFile.WriteString(fmt.Sprintf(downloadProductConfigWithVariablesTmpl, tempDir))
					Expect(err).NotTo(HaveOccurred())

					err = configFile.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err = os.RemoveAll(configFile.Name())
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error if missing variables", func() {
					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
				})

				Context("passed in a vars-file", func() {
					var varsFile *os.File

					BeforeEach(func() {
						varsFile, err = ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())

						_, err = varsFile.WriteString(`product-slug: elastic-runtime`)
						Expect(err).NotTo(HaveOccurred())

						err = varsFile.Close()
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						err = os.RemoveAll(varsFile.Name())
						Expect(err).NotTo(HaveOccurred())
					})

					It("can interpolate variables into the configuration", func() {
						err = command.Execute([]string{
							"--config", configFile.Name(),
							"--vars-file", varsFile.Name(),
						})
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("passed as environment variables", func() {
					BeforeEach(func() {
						environFunc = func() []string {
							return []string{"OM_VAR_product-slug='sea-slug'"}
						}
					})

					It("can interpolate variables into the configuration", func() {
						err = command.Execute([]string{
							"--config", configFile.Name(),
							"--vars-env", "OM_VAR",
						})
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})

		Describe("managing and reporting the filename written to the filesystem", func() {
			When("S3 configuration is provided and, blobstore is not set", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/my-great-product.pivotal",
							Name:         "my-great-product.pivotal",
						},
					}, nil)

					commandArgs = []string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
						"--s3-bucket", "there once was a man from a",
					}
				})

				It("prefixes the filename with a bracketed slug and version", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					prefixedFileName := path.Join(tempDir, "[mayhem-crew,2.0.0]my-great-product.pivotal")
					Expect(prefixedFileName).To(BeAnExistingFile())
				})

				It("writes the prefixed filename to the download-file.json", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
					Expect(err).NotTo(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					prefixedFileName := path.Join(tempDir, "[mayhem-crew,2.0.0]my-great-product.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "mayhem-crew" }`, prefixedFileName)))
				})
			})

			When("S3 configuration is not provided, and blobstore is not set", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/my-great-product.pivotal",
							Name:         "my-great-product.pivotal",
						},
					}, nil)

					commandArgs = []string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					}
				})
				It("doesn't prefix", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					unPrefixedFileName := path.Join(tempDir, "my-great-product.pivotal")
					Expect(unPrefixedFileName).To(BeAnExistingFile())
				})

				It("writes the unprefixed filename to the download-file.json", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
					Expect(err).NotTo(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					unPrefixedFileName := path.Join(tempDir, "my-great-product.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "mayhem-crew" }`, unPrefixedFileName)))
				})
			})

		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				err = command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse download-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when a required flag is not provided", func() {
			It("returns an error", func() {
				err = command.Execute([]string{})
				Expect(err).To(MatchError("could not parse download-product flags: missing required flag \"--output-directory\""))
			})
		})

		Context("when both product-version and product-version-regex are set", func() {
			It("fails with an error saying that the user must pick one or the other", func() {
				err = command.Execute([]string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--product-version-regex", ".*",
					"--output-directory", tempDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot use both --product-version and --product-version-regex; please choose one or the other"))
			})
		})

		Context("when neither product-version nor product-version-regex are set", func() {
			It("fails with an error saying that the user must provide one or the other", func() {
				err = command.Execute([]string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--output-directory", tempDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no version information provided; please provide either --product-version or --product-version-regex"))
			})
		})

		Context("when the release specified is not available", func() {
			BeforeEach(func() {
				fakePivnetDownloader.ReleaseForVersionReturns(pivnet.Release{}, fmt.Errorf("some-error"))
			})

			It("returns an error", func() {
				err = command.Execute([]string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not fetch the release for elastic-runtime 2.0.0: some-error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
				ShortDescription: "downloads a specified product file from Pivotal Network",
				Flags:            command.Options,
			}))
		})
	})
})
