package commands_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/go-pivnet"
	log "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/validator"
)

var _ = Describe("DownloadProduct", func() {
	var (
		callCount            int
		command              commands.DownloadProduct
		commandArgs          []string
		logger               *loggerfakes.FakeLogger
		fakePivnetDownloader *fakes.PivnetDownloader
		fakeStower           *mockStower
		fakeWriter           *gbytes.Buffer
		environFunc          func() []string
		tempDir              string
		err                  error
	)

	fakePivnetFactory := func(config pivnet.ClientConfig, logger log.Logger) commands.PivnetDownloader {
		return fakePivnetDownloader
	}

	BeforeEach(func() {
		callCount = 0
		logger = &loggerfakes.FakeLogger{}
		fakePivnetDownloader = &fakes.PivnetDownloader{}
		fakeStower = newMockStower([]mockItem{}, &callCount)
		environFunc = func() []string { return nil }
		fakeWriter = gbytes.NewBuffer()
	})

	JustBeforeEach(func() {
		command = commands.NewDownloadProduct(environFunc, logger, fakeWriter, fakePivnetFactory, fakeStower)
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
			Expect(*fakeStower.dialCallCount).To(Equal(0))
		})

		When("the blobstore flag is set to s3", func() {
			BeforeEach(func() {
				commandArgs = []string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
					"--blobstore", "s3",
					"--s3-config", `
bucket: bucket
access-key-id: access-key-id
secret-access-key: secret-access-key
region-name: region-name
endpoint: endpoint
`,
				}
			})

			It("downloads the specified product from s3", func() {
				command.Execute(commandArgs)

				Expect(*fakeStower.dialCallCount).Should(BeNumerically(">", 0))
				Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(0))
			})
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
				Expect(file.Name()).To(Equal(path.Join(tempDir, "elastic-runtime-2.1.2_cf-2.1-build.11.pivotal")))
				Expect(slug).To(Equal("elastic-runtime"))
				Expect(releaseID).To(Equal(4))
				Expect(productFileID).To(Equal(54321))

				prefixedFileName := path.Join(tempDir, "elastic-runtime-2.1.2_cf-2.1-build.11.pivotal")
				Expect(prefixedFileName).To(BeAnExistingFile())
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

		Context("when the download-stemcell flag is set", func() {
			BeforeEach(func() {
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
				err = command.Execute([]string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
					"--stemcell-iaas", "google",
				})
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

				fileName := path.Join(tempDir, commands.DownloadProductOutputFilename)
				fileContent, err := ioutil.ReadFile(fileName)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileName).To(BeAnExistingFile())
				prefixedFileName := path.Join(tempDir, "elastic-runtime-2.0.0_cf-2.0-build.1.pivotal")
				Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`
					{
						"product_path": "%s",
						"product_slug": "elastic-runtime",
						"stemcell_path": "%s",
						"stemcell_version": "97.19"
					}`, prefixedFileName, stemcellFile.Name())))
			})

			Context("when the product is not a tile and download-stemcell flag is set", func() {
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

				It("exit gracefully when the product is not a tile", func() {
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
			BeforeEach(func() {
				filePath := path.Join(tempDir, "elastic-runtime-2.0.0_cf-2.0-build.1.pivotal")
				file, err := os.Create(filePath)
				Expect(err).NotTo(HaveOccurred())
				_, err = file.WriteString("something-not-important")
				Expect(err).NotTo(HaveOccurred())
				err = file.Close()
				Expect(err).NotTo(HaveOccurred())

				validator := validator.NewSHA256Calculator()
				sum, err := validator.Checksum(filePath)
				Expect(err).NotTo(HaveOccurred())

				fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
					{
						ID:           54321,
						AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.pivotal",
						SHA256:       sum,
						Name:         "Example Cloud Foundry",
					},
				}, nil)
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
			When("the file to be downloaded already satisfies our blobstore parsability constraints", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1-electric-teeth-2.0.0.pivotal",
							Name:         "cf-2.0-build.1.pivotal",
						},
					}, nil)

					commandArgs = []string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "electric-teeth",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					}
				})

				It("does not duplicate the filename prefix in the output file or filename", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					prefixedFileName := path.Join(tempDir, "cf-2.0-build.1-electric-teeth-2.0.0.pivotal")
					Expect(prefixedFileName).To(BeAnExistingFile())
				})
				It("writes the un-modified filname in the download-file.json", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, commands.DownloadProductOutputFilename)
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
					Expect(err).NotTo(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					prefixedFileName := path.Join(tempDir, "cf-2.0-build.1-electric-teeth-2.0.0.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "electric-teeth" }`, prefixedFileName)))
				})
			})

			When("the slug comes after the version in the name of the file to be downloaded", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/2.0.0_downtown-buzz-edition-cf-2.0-build.1.pivotal",
							Name:         "cf-2.0-build.1.pivotal",
						},
					}, nil)

					commandArgs = []string{
						"--pivnet-api-token", "token",
						"--pivnet-file-glob", "*.pivotal",
						"--pivnet-product-slug", "downtown-buzz-edition",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					}
				})

				It("prefixes the filename with slug and version to satisfy our parsing constraints", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					prefixedFileName := path.Join(tempDir, "downtown-buzz-edition-2.0.0_2.0.0_downtown-buzz-edition-cf-2.0-build.1.pivotal")
					Expect(prefixedFileName).To(BeAnExistingFile())
				})

				It("writes the prefixed filename in the download-file.json", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, commands.DownloadProductOutputFilename)
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
					Expect(err).NotTo(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					prefixedFileName := path.Join(tempDir, "downtown-buzz-edition-2.0.0_2.0.0_downtown-buzz-edition-cf-2.0-build.1.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "downtown-buzz-edition" }`, prefixedFileName)))
				})
			})

			When("the slug does not exist in the name of the file to be downloaded", func() {
				BeforeEach(func() {
					fakePivnetDownloader.ProductFilesForReleaseReturnsOnCall(0, []pivnet.ProductFile{
						{
							ID:           54321,
							AWSObjectKey: "/some-account/some-bucket/2.0.0-cf-2.0-build.1.pivotal",
							Name:         "cf-2.0-build.1.pivotal",
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

				It("prefixes the filename with slug and version to satisfy our parsing constraints", func() {
					err = command.Execute(commandArgs)
					Expect(err).NotTo(HaveOccurred())

					prefixedFileName := path.Join(tempDir, "mayhem-crew-2.0.0_2.0.0-cf-2.0-build.1.pivotal")
					Expect(prefixedFileName).To(BeAnExistingFile())
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
