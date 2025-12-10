package commands_test

import (
	"archive/zip"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/pivotal-cf/om/extractor"

	"github.com/pivotal-cf/om/download_clients"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/om/commands"
	cmdFakes "github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/download_clients/fakes"
)

var _ = Describe("DownloadProduct", func() {
	var (
		command                    *commands.DownloadProduct
		environFunc                func() []string
		err                        error
		fakeProductDownloader      *fakes.ProductDownloader
		fakeDownloadProductService *cmdFakes.DownloadProductService
		buffer                     *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeProductDownloader = &fakes.ProductDownloader{}
		fakeDownloadProductService = &cmdFakes.DownloadProductService{}
		fakeProductDownloader.GetAllProductVersionsReturns([]string{"2.0.0"}, nil)
		environFunc = func() []string { return nil }
	})

	JustBeforeEach(func() {
		download_clients.NewPivnetClient = func(stdout *log.Logger, stderr *log.Logger, factory download_clients.PivnetFactory, token string, skipSSL bool, pivnetHost string, proxyURL string, proxyUsername string, proxyPassword string, proxyAuthType string, proxyKrb5Config string) (download_clients.ProductDownloader, error) {
			return fakeProductDownloader, nil
		}
		buffer = gbytes.NewBuffer()
		command = commands.NewDownloadProduct(environFunc, log.New(buffer, "", 0), log.New(buffer, "", 0), buffer, fakeDownloadProductService)
	})

	When("the flags are set correctly", func() {
		When("it can connect to the source", func() {
			BeforeEach(func() {
				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")

				fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
			})

			It("downloads a product from the downloader", func() {
				tempDir, err := os.MkdirTemp("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				}

				err = executeCommand(command, commandArgs)
				Expect(err).ToNot(HaveOccurred())
			})

			It("supports the pivnet-file-glob alias for file-glob", func() {
				tempDir, err := os.MkdirTemp("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				}

				err = executeCommand(command, commandArgs)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("a valid product-version-regex is provided", func() {
			BeforeEach(func() {
				fakeProductDownloader.GetAllProductVersionsReturns(
					[]string{"3.0.0", "1.1.11", "2.1.2", "2.1.1", "2.0.1"},
					nil,
				)
				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.1-build.11.pivotal")
				fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)
			})

			It("downloads the highest version matching that regex", func() {
				tempDir, err := os.MkdirTemp("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version-regex", `2\..\..*`,
					"--output-directory", tempDir,
				}

				err = executeCommand(command, commandArgs)
				Expect(err).ToNot(HaveOccurred())

				slug := fakeProductDownloader.GetAllProductVersionsArgsForCall(0)
				Expect(slug).To(Equal("elastic-runtime"))

				slug, version, _ := fakeProductDownloader.GetLatestProductFileArgsForCall(0)
				Expect(slug).To(Equal("elastic-runtime"))
				Expect(version).To(Equal("2.1.2"))

				_, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(0)
				Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "cf-2.1-build.11.pivotal.partial")))

				Expect(filepath.Join(tempDir, "cf-2.1-build.11.pivotal")).To(BeAnExistingFile())
				Expect(filepath.Join(tempDir, "cf-2.1-build.11.pivotal.partial")).ToNot(BeAnExistingFile())
			})

			When("the releases contains non-semver-compatible version", func() {
				BeforeEach(func() {
					fakeProductDownloader.GetAllProductVersionsReturns(
						[]string{"2.1.2", "2.0.x"},
						nil,
					)
				})

				It("ignores the version and prints a warning", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version-regex", `2\..\..*`,
						"--output-directory", tempDir,
					}

					err = executeCommand(command, commandArgs)
					Expect(err).ToNot(HaveOccurred())

					Eventually(buffer).Should(gbytes.Say("warning: could not parse semver version from: 2.0.x"))
				})
			})

			When("there are no valid versions found for given product regex", func() {
				BeforeEach(func() {
					fakeProductDownloader.GetAllProductVersionsReturns(
						[]string{"3.1.2"},
						nil,
					)
				})

				It("returns an error", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version-regex", `2\..\..*`,
						"--output-directory", tempDir,
					}

					err = executeCommand(command, commandArgs)
					Expect(err).To(MatchError(ContainSubstring("no valid versions found for product \"elastic-runtime\"")))
				})
			})
		})

		When("a file is being downloaded with a SHA sum value from the downloader", func() {
			When("the shasum is valid for the downloaded file", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fa.SHA256Returns("d1b2a59fbea7e20077af9f91b27e95e865061b270be03ff539ab3b73587882e8")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						return os.WriteFile(file.Name(), []byte("contents"), 0777)
					}
				})

				It("downloads a product from the downloader", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					}

					err = executeCommand(command, commandArgs)
					Expect(err).ToNot(HaveOccurred())
					Expect(filepath.Join(tempDir, "cf-2.0-build.1.pivotal")).To(BeAnExistingFile())
				})
			})

			When("the shasum is invalid for the downloaded file", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fa.SHA256Returns("asdfasdf")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						return os.WriteFile(file.Name(), []byte("contents"), 0777)
					}
				})

				It("errors and removes the file from the file system", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					}

					err = executeCommand(command, commandArgs)
					Expect(err).To(HaveOccurred())
					Expect(filepath.Join(tempDir, "cf-2.0-build.1.pivotal")).ToNot(BeAnExistingFile())
				})
			})
		})

		When("the stemcell-iaas flag is set", func() {
			When("the product has an associated stemcell", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fa.ProductMetadataReturns(&extractor.Metadata{Name: "fake-tile", Version: "2.0.0"}, nil)
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fa = &fakes.FileArtifacter{}
					fa.NameReturns("stemcell.tgz")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, fa, nil)

					sa := &fakes.StemcellArtifacter{}
					sa.SlugReturns("stemcells-ubuntu-xenial")
					sa.VersionReturns("97.190")
					fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)
				})

				It("grabs the latest stemcell for the product that matches the glob", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						createProductPivotalFile(file)
						return nil
					}

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "google",
					}

					err = executeCommand(command, commandArgs)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeProductDownloader.GetLatestStemcellForProductCallCount()).To(Equal(1))
					Expect(fakeProductDownloader.GetLatestProductFileCallCount()).To(Equal(2))
					Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(2))
					Expect(fakeProductDownloader.GetAllProductVersionsCallCount()).To(Equal(1))

					fa, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(1)
					Expect(fa.Name()).To(Equal("stemcell.tgz"))
					Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "stemcell.tgz.partial")))

					fileName := path.Join(tempDir, "download-file.json")
					fileContent, err := os.ReadFile(fileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(fileName).To(BeAnExistingFile())
					downloadedFilePath := path.Join(tempDir, "cf-2.0-build.1.pivotal")
					downloadedStemcellFilePath := path.Join(tempDir, "stemcell.tgz")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`
							{
								"product_path": "%s",
								"product_slug": "elastic-runtime",
								"product_version": "2.0.0",
								"stemcell_path": "%s",
								"stemcell_version": "97.190"
							}`, downloadedFilePath, downloadedStemcellFilePath)))

					fileName = path.Join(tempDir, "assign-stemcell.yml")
					fileContent, err = os.ReadFile(fileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(fileName).To(BeAnExistingFile())
					Expect(string(fileContent)).To(MatchJSON(`
							{
								"product": "fake-tile",
								"stemcell": "97.190"
							}`))
				})

				When("the --check-upload-already is specified", func() {
					It("does not download the stemcell and product", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.pivotal",
							"--pivnet-product-slug", "elastic-runtime",
							"--product-version", "2.0.0",
							"--output-directory", tempDir,
							"--stemcell-iaas", "google",
							"--check-already-uploaded",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).ToNot(HaveOccurred())

						assignStemcellFilename := path.Join(tempDir, "assign-stemcell.yml")
						Expect(assignStemcellFilename).To(BeAnExistingFile())
					})
				})

				When("the --stemcell-output-dir flag is passed", func() {
					var (
						commandArgs       []string
						productOutputDir  string
						stemcellOutputDir string
					)

					tempFile := func(dir, pattern string) string {
						file, err := os.CreateTemp(dir, pattern)
						Expect(err).ToNot(HaveOccurred())
						return file.Name()
					}

					BeforeEach(func() {
						productOutputDir, err = os.MkdirTemp("", "om-tests-output-dir-")
						Expect(err).ToNot(HaveOccurred())

						stemcellOutputDir, err = os.MkdirTemp("", "om-tests-stemcell-output-dir-")
						Expect(err).ToNot(HaveOccurred())

						fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
							createProductPivotalFile(file)
							return nil
						}

						commandArgs = []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.pivotal",
							"--pivnet-product-slug", "elastic-runtime",
							"--product-version", "2.0.0",
							"--output-directory", productOutputDir,
							"--stemcell-output-directory", stemcellOutputDir,
							"--stemcell-iaas", "google",
						}
					})

					It("downloads the product and stemcell to their specified directories", func() {
						err = executeCommand(command, commandArgs)
						Expect(err).ToNot(HaveOccurred())

						downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
						downloadedStemcellFilePath := path.Join(stemcellOutputDir, "stemcell.tgz")
						Expect(downloadedFilePath).To(BeAnExistingFile())
						Expect(downloadedStemcellFilePath).To(BeAnExistingFile())
					})

					When("CACHE_CLEANUP is an invalid value", func() {
						BeforeEach(func() {
							os.Setenv("CACHE_CLEANUP", "invalid")
						})

						AfterEach(func() {
							os.Unsetenv("CACHE_CLEANUP")
						})

						It("does not cleanup the cache", func() {
							alreadyDownloadedProduct := tempFile(productOutputDir, "product*.pivotal")
							alreadyDownloadedLightStemcell := tempFile(stemcellOutputDir, "light-bosh-google-*.tgz")
							unknownFileWeDontOwn := tempFile(productOutputDir, "no-delete")

							err = executeCommand(command, commandArgs)
							Expect(err).ToNot(HaveOccurred())

							downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
							downloadedStemcellFilePath := path.Join(stemcellOutputDir, "stemcell.tgz")
							Expect(downloadedFilePath).To(BeAnExistingFile())
							Expect(downloadedStemcellFilePath).To(BeAnExistingFile())

							Expect(alreadyDownloadedProduct).To(BeAnExistingFile())
							Expect(alreadyDownloadedLightStemcell).To(BeAnExistingFile())
							Expect(unknownFileWeDontOwn).To(BeAnExistingFile())
						})
					})

					When("CACHE_CLEANUP='I acknowledge this will delete files in the output directories' is passed along with both output-dir flags", func() {
						BeforeEach(func() {
							os.Setenv("CACHE_CLEANUP", "I acknowledge this will delete files in the output directories")
						})

						AfterEach(func() {
							os.Unsetenv("CACHE_CLEANUP")
						})

						It("only deletes files that match the glob of the product and stemcell(s)", func() {
							alreadyDownloadedProduct := tempFile(productOutputDir, "product*.pivotal")
							alreadyDownloadedLightStemcell := tempFile(stemcellOutputDir, "light-bosh-google-*.tgz")
							unknownFileWeDontOwn := tempFile(productOutputDir, "no-delete")

							err = executeCommand(command, commandArgs)
							Expect(err).ToNot(HaveOccurred())

							downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
							downloadedStemcellFilePath := path.Join(stemcellOutputDir, "stemcell.tgz")
							Expect(downloadedFilePath).To(BeAnExistingFile())
							Expect(downloadedStemcellFilePath).To(BeAnExistingFile())

							Expect(alreadyDownloadedProduct).ToNot(BeAnExistingFile())
							Expect(alreadyDownloadedLightStemcell).ToNot(BeAnExistingFile())
							Expect(unknownFileWeDontOwn).To(BeAnExistingFile())
						})

						When("the product and stemcell have already been downloaded and cached", func() {
							It("only deletes previous versions of the product", func() {
								fa := &fakes.FileArtifacter{}
								fa.NameReturns("light-bosh-google-2-stemcell.tgz")
								fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, fa, nil)

								previousDownloadedProduct := tempFile(productOutputDir, "cf-2.0-build.*.pivotal")
								previousDownloadedStemcell := tempFile(stemcellOutputDir, "light-bosh-google-1-stemcell*.tgz")
								downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
								downloadedStemcellFilePath := path.Join(stemcellOutputDir, "light-bosh-google-2-stemcell.tgz")
								downloadedFile, err := os.Create(downloadedFilePath)
								Expect(err).ToNot(HaveOccurred())
								createProductPivotalFile(downloadedFile)

								err = executeCommand(command, commandArgs)
								Expect(err).ToNot(HaveOccurred())

								Expect(downloadedFilePath).To(BeAnExistingFile())
								Expect(downloadedStemcellFilePath).To(BeAnExistingFile())

								Expect(previousDownloadedProduct).ToNot(BeAnExistingFile())
								Expect(previousDownloadedStemcell).ToNot(BeAnExistingFile())
							})
						})
					})
				})

				When("the --stemcell-version flag is passed", func() {
					It("downloads the specified stemcell at that version", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
							createProductPivotalFile(file)
							return nil
						}

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.pivotal",
							"--pivnet-product-slug", "elastic-runtime",
							"--product-version", "2.0.0",
							"--output-directory", tempDir,
							"--stemcell-iaas", "google",
							"--stemcell-version", "100.00",
							"--s3-bucket", "there once was a man from a",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).ToNot(HaveOccurred())

						fa, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(1)
						Expect(fa.Name()).To(Equal("stemcell.tgz"))
						Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "[stemcells-ubuntu-xenial,100.00]stemcell.tgz.partial")))

						fileName := path.Join(tempDir, "download-file.json")
						fileContent, err := os.ReadFile(fileName)
						Expect(err).ToNot(HaveOccurred())
						Expect(fileName).To(BeAnExistingFile())
						downloadedFilePath := path.Join(tempDir, "[elastic-runtime,2.0.0]cf-2.0-build.1.pivotal")
						downloadedStemcellFilePath := path.Join(tempDir, "[stemcells-ubuntu-xenial,100.00]stemcell.tgz")
						Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`
							{
								"product_path": "%s",
								"product_slug": "elastic-runtime",
								"product_version": "2.0.0",
								"stemcell_path": "%s",
								"stemcell_version": "100.00"
							}`, downloadedFilePath, downloadedStemcellFilePath)))

						fileName = path.Join(tempDir, "assign-stemcell.yml")
						fileContent, err = os.ReadFile(fileName)
						Expect(err).ToNot(HaveOccurred())
						Expect(fileName).To(BeAnExistingFile())
						Expect(string(fileContent)).To(MatchJSON(`
							{
								"product": "fake-tile",
								"stemcell": "100.00"
							}`))
					})
				})
			})

			When("the product is not a tile", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.tgz")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)
				})

				It("prints a warning and returns available file artifacts", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "google",
					})

					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := os.ReadFile(downloadReportFileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					downloadedFilePath := path.Join(tempDir, "cf-2.0-build.1.tgz")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`
							{
								"product_path": "%s",
								"product_slug": "elastic-runtime",
								"product_version": "2.0.0"
							}`, downloadedFilePath)))
					Expect(buffer).Should(gbytes.Say("the downloaded file is not a .pivotal file. Not determining and fetching required stemcell."))
				})
			})

			When("the stemcell cannot be downloaded", func() {
				It("returns an error message", func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, nil, errors.New("some error"))
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(2, nil, errors.New("some error"))

					sa := &fakes.StemcellArtifacter{}
					sa.SlugReturns("stemcells-ubuntu-xenial")
					sa.VersionReturns("97.190")
					fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)

					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						createProductPivotalFile(file)
						return nil
					}

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "unknown-iaas",
					}

					err = executeCommand(command, commandArgs)
					Expect(err).To(MatchError(ContainSubstring("No stemcell identified for IaaS \"unknown-iaas\" on Pivotal Network. Correct the `stemcell-iaas` option to match the IaaS portion of the stemcell filename, or remove the option")))
				})
			})

			When("the --stemcell-heavy flag is not provided", func() {
				It("downloads the corresponding light or heavy stemcell", func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fa.ProductMetadataReturns(&extractor.Metadata{}, nil)
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						createProductPivotalFile(file)
						return nil
					}

					fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, nil, errors.New("some-error"))
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(2, &fakes.FileArtifacter{}, nil)

					sa := &fakes.StemcellArtifacter{}
					sa.SlugReturns("stemcells-ubuntu-xenial")
					sa.VersionReturns("97.190")
					fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)

					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "aws",
					})

					Expect(err).ToNot(HaveOccurred())
					Expect(fakeProductDownloader.GetLatestProductFileCallCount()).To(Equal(3))
					_, _, glob := fakeProductDownloader.GetLatestProductFileArgsForCall(1)
					Expect(glob).To(Equal("light*bosh*aws*"))

					_, _, glob = fakeProductDownloader.GetLatestProductFileArgsForCall(2)
					Expect(glob).To(Equal("bosh*aws*"))
				})
			})

			When("providing the --stemcell-heavy flag", func() {
				It("downloads the corresponding light or heavy stemcell", func() {
					fa := &fakes.FileArtifacter{}
					fa.ProductMetadataReturns(&extractor.Metadata{}, nil)
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						createProductPivotalFile(file)
						return nil
					}

					fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, &fakes.FileArtifacter{}, nil)

					sa := &fakes.StemcellArtifacter{}
					sa.SlugReturns("stemcells-ubuntu-xenial")
					sa.VersionReturns("97.190")
					fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)

					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "aws",
						"--stemcell-heavy",
					})

					Expect(err).ToNot(HaveOccurred())

					Expect(fakeProductDownloader.GetLatestProductFileCallCount()).To(Equal(2))
					_, _, glob := fakeProductDownloader.GetLatestProductFileArgsForCall(1)
					Expect(glob).To(Equal("bosh*aws*"))
				})

				It("fails if --stemcell-heavy flag is provided but --stemcell-iaas is not", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-heavy",
					})

					Expect(err).To(MatchError(ContainSubstring("--stemcell-heavy requires --stemcell-iaas to be defined")))
				})

				It("fails is --stemcell-heavy is provided but the heavy stemcell does not exist", func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
					fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

					fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
						createProductPivotalFile(file)
						return nil
					}

					fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, nil, errors.New("no stemcell"))

					sa := &fakes.StemcellArtifacter{}
					sa.SlugReturns("stemcells-ubuntu-xenial")
					sa.VersionReturns("97.190")
					fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)

					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "aws",
						"--stemcell-heavy",
					})

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("No heavy stemcell identified for IaaS \"aws\" on Pivotal Network"))
				})
			})
		})

		When("the --check-already-uploaded is set", func() {
			When("looking up a stemcell as a product", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/stemcell.tgz")
					fa.ProductMetadataReturns(&extractor.Metadata{Name: "xenial-stemcells", Version: "100.0"}, nil)
					fakeProductDownloader.GetAllProductVersionsReturns([]string{"100.0"}, nil)

					fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
				})

				When("the stemcell already exists on the OpsManager", func() {
					BeforeEach(func() {
						fakeDownloadProductService.CheckStemcellAvailabilityReturns(true, nil)
					})

					It("does not download it", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.tgz",
							"--pivnet-product-slug", "xenial-stemcells",
							"--product-version", "100.0",
							"--output-directory", tempDir,
							"--check-already-uploaded",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(0))
						Expect(fakeDownloadProductService.CheckStemcellAvailabilityCallCount()).To(Equal(1))
						filename := fakeDownloadProductService.CheckStemcellAvailabilityArgsForCall(0)
						Expect(filename).To(Equal("stemcell.tgz"))
					})
				})

				When("the stemcell is not on the OpsManager", func() {
					BeforeEach(func() {
						fakeDownloadProductService.CheckStemcellAvailabilityReturns(false, nil)
					})

					It("download the file", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.tgz",
							"--pivnet-product-slug", "xenial-stemcells",
							"--product-version", "100.0",
							"--output-directory", tempDir,
							"--check-already-uploaded",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(1))
						_, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(0)
						Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "stemcell.tgz.partial")))
						Expect(filepath.Join(tempDir, "stemcell.tgz")).To(BeAnExistingFile())

						Expect(fakeDownloadProductService.CheckStemcellAvailabilityCallCount()).To(Equal(1))
						filename := fakeDownloadProductService.CheckStemcellAvailabilityArgsForCall(0)
						Expect(filename).To(Equal("stemcell.tgz"))
					})
				})

				When("the stemcell check service returns an error", func() {
					BeforeEach(func() {
						fakeDownloadProductService.CheckStemcellAvailabilityReturns(false, errors.New("some error"))
					})

					It("returns that error", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.tgz",
							"--pivnet-product-slug", "xenial-stemcells",
							"--product-version", "100.0",
							"--output-directory", tempDir,
							"--check-already-uploaded",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("some error"))

						Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(0))
						Expect(fakeDownloadProductService.CheckStemcellAvailabilityCallCount()).To(Equal(1))
					})
				})
			})

			When("looking up a pivotal file", func() {
				When("the metadata cannot be read", func() {
					BeforeEach(func() {
						fa := &fakes.FileArtifacter{}
						fa.NameReturns("/some-account/some-bucket/cf-2.1-build.11.pivotal")
						fa.ProductMetadataReturns(nil, errors.New("some error"))

						fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
					})

					It("does not download it", func() {
						tempDir, err := os.MkdirTemp("", "om-tests-")
						Expect(err).ToNot(HaveOccurred())

						commandArgs := []string{
							"--pivnet-api-token", "token",
							"--file-glob", "*.pivotal",
							"--pivnet-product-slug", "elastic-runtime",
							"--product-version", "2.0.0",
							"--output-directory", tempDir,
							"--check-already-uploaded",
						}

						err = executeCommand(command, commandArgs)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("some error"))
					})
				})

				When("the metadata can be found", func() {
					BeforeEach(func() {
						fa := &fakes.FileArtifacter{}
						fa.NameReturns("/some-account/some-bucket/cf-2.1-build.11.pivotal")
						fa.ProductMetadataReturns(&extractor.Metadata{Name: "example-product", Version: "1.2.3"}, nil)

						fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
					})

					When("the product already exists on the OpsManager", func() {
						BeforeEach(func() {
							fakeDownloadProductService.CheckProductAvailabilityReturns(true, nil)
						})

						It("does not download it", func() {
							tempDir, err := os.MkdirTemp("", "om-tests-")
							Expect(err).ToNot(HaveOccurred())

							commandArgs := []string{
								"--pivnet-api-token", "token",
								"--file-glob", "*.pivotal",
								"--pivnet-product-slug", "elastic-runtime",
								"--product-version", "2.0.0",
								"--output-directory", tempDir,
								"--check-already-uploaded",
							}

							err = executeCommand(command, commandArgs)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(0))
							Expect(fakeDownloadProductService.CheckProductAvailabilityCallCount()).To(Equal(1))
							name, version := fakeDownloadProductService.CheckProductAvailabilityArgsForCall(0)
							Expect(name).To(Equal("example-product"))
							Expect(version).To(Equal("1.2.3"))
						})
					})

					When("the product is not on the OpsManager", func() {
						BeforeEach(func() {
							fakeDownloadProductService.CheckProductAvailabilityReturns(false, nil)
						})

						It("download the file", func() {
							tempDir, err := os.MkdirTemp("", "om-tests-")
							Expect(err).ToNot(HaveOccurred())

							commandArgs := []string{
								"--pivnet-api-token", "token",
								"--file-glob", "*.pivotal",
								"--pivnet-product-slug", "elastic-runtime",
								"--product-version", "2.0.0",
								"--output-directory", tempDir,
								"--check-already-uploaded",
							}

							err = executeCommand(command, commandArgs)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(1))
							_, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(0)
							Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "cf-2.1-build.11.pivotal.partial")))
							Expect(filepath.Join(tempDir, "cf-2.1-build.11.pivotal")).To(BeAnExistingFile())

							Expect(fakeDownloadProductService.CheckProductAvailabilityCallCount()).To(Equal(1))
							name, version := fakeDownloadProductService.CheckProductAvailabilityArgsForCall(0)

							Expect(name).To(Equal("example-product"))
							Expect(version).To(Equal("1.2.3"))
						})
					})

					When("the product check service returns an error", func() {
						BeforeEach(func() {
							fakeDownloadProductService.CheckProductAvailabilityReturns(false, errors.New("some error"))
						})

						It("returns that error", func() {
							tempDir, err := os.MkdirTemp("", "om-tests-")
							Expect(err).ToNot(HaveOccurred())

							commandArgs := []string{
								"--pivnet-api-token", "token",
								"--file-glob", "*.pivotal",
								"--pivnet-product-slug", "elastic-runtime",
								"--product-version", "2.0.0",
								"--output-directory", tempDir,
								"--check-already-uploaded",
							}

							err = executeCommand(command, commandArgs)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("some error"))

							Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(0))
							Expect(fakeDownloadProductService.CheckProductAvailabilityCallCount()).To(Equal(1))
						})
					})
				})
			})
		})

		When("the product is already downloaded", func() {
			var tempDir string

			BeforeEach(func() {
				tempDir, err = os.MkdirTemp("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
				fa.ProductMetadataReturns(&extractor.Metadata{}, nil)
				fakeProductDownloader.GetLatestProductFileReturnsOnCall(0, fa, nil)

				fa = &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/light-bosh-stemcell-97.19-google-kvm-ubuntu-xenial-go_agent.tgz")
				fakeProductDownloader.GetLatestProductFileReturnsOnCall(1, fa, nil)

				sa := &fakes.StemcellArtifacter{}
				sa.SlugReturns("stemcells-ubuntu-xenial")
				sa.VersionReturns("97.19")
				fakeProductDownloader.GetLatestStemcellForProductReturns(sa, nil)

				fakeProductDownloader.DownloadProductToFileStub = func(artifacter download_clients.FileArtifacter, file *os.File) error {
					createProductPivotalFile(file)
					return nil
				}

				filePath := path.Join(tempDir, "cf-2.0-build.1.pivotal")
				file, err := os.Create(filePath)
				Expect(err).ToNot(HaveOccurred())
				createProductPivotalFile(file)
			})

			It("does not re-download the product", func() {
				err = executeCommand(command, []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(0))
				Expect(buffer).Should(gbytes.Say("already exists, skip downloading"))
			})

			It("still downloads the stemcell if not already downloaded", func() {
				err = executeCommand(command, []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--stemcell-iaas", "google",
					"--output-directory", tempDir,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(1))
			})
		})

		Describe("managing and reporting the filename written to the filesystem", func() {
			When("S3 configuration is provided and source is not set", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/my-great-product.pivotal")
					fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
				})

				It("prefixes the filename with a bracketed slug and version", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
						"--s3-bucket", "there once was a man from a",
					})
					Expect(err).ToNot(HaveOccurred())

					prefixedFileName := path.Join(tempDir, "[mayhem-crew,2.0.0]my-great-product.pivotal")
					Expect(prefixedFileName).To(BeAnExistingFile())
				})

				It("writes the prefixed filename to the download-file.json", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
						"--s3-bucket", "there once was a man from a",
					})
					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := os.ReadFile(downloadReportFileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					prefixedFileName := path.Join(tempDir, "[mayhem-crew,2.0.0]my-great-product.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "mayhem-crew", "product_version": "2.0.0" }`, prefixedFileName)))
				})
			})

			When("S3 configuration is not provided and source is not set", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/my-great-product.pivotal")
					fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
				})
				It("doesn't prefix", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())

					unPrefixedFileName := path.Join(tempDir, "my-great-product.pivotal")
					Expect(unPrefixedFileName).To(BeAnExistingFile())
				})

				It("writes the unprefixed filename to the download-file.json", func() {
					tempDir, err := os.MkdirTemp("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := os.ReadFile(downloadReportFileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(downloadReportFileName).To(BeAnExistingFile())
					unPrefixedFileName := path.Join(tempDir, "my-great-product.pivotal")
					Expect(string(fileContent)).To(MatchJSON(fmt.Sprintf(`{"product_path": "%s", "product_slug": "mayhem-crew", "product_version": "2.0.0" }`, unPrefixedFileName)))
				})
			})
		})
	})

	When("--stemcell-version flag is provided, but --stemcell-iaas is missing", func() {
		It("returns an error", func() {
			tempDir, err := os.MkdirTemp("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--output-directory", tempDir,
				"--stemcell-version", "100.0",
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("--stemcell-version requires --stemcell-iaas to be defined")))
		})
	})

	When("directory flags are provided pointing to directories that don't exist", func() {
		var (
			nonexistingDir string
			validDirectory string
			err            error
		)

		BeforeEach(func() {
			nonexistingDir = "/invalid/dir/noexist"
			validDirectory, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors, printing both --output-dir and the filepath in question", func() {
			err := executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--stemcell-output-directory", validDirectory,
				"--stemcell-iaas", "aws",
				"--output-directory", nonexistingDir,
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`--output-directory "/invalid/dir/noexist" does not exist: open /invalid/dir/noexist: no such file or directory`))
		})

		It("errors, printing both --stemcell-output-dir and the filepath in question", func() {
			err := executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--stemcell-output-directory", nonexistingDir,
				"--stemcell-iaas", "aws",
				"--output-directory", validDirectory,
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`--stemcell-output-directory "/invalid/dir/noexist" does not exist: open /invalid/dir/noexist: no such file or directory`))
		})
	})

	When("directory flags are provided pointing to non-directory files", func() {
		var (
			existingNonDirFile *os.File
			validDirectory     string
			err                error
		)

		BeforeEach(func() {
			existingNonDirFile, err = os.CreateTemp("", "")
			Expect(err).NotTo(HaveOccurred())

			validDirectory, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors, printing both --output-directory and the filepath in question", func() {
			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--stemcell-output-directory", validDirectory,
				"--stemcell-iaas", "aws",
				"--output-directory", existingNonDirFile.Name(),
			})

			expectedOutput := fmt.Sprintf("--output-directory %q is not a directory", existingNonDirFile.Name())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedOutput))
		})

		It("errors, printing both --stemcell-output-directory and the filepath in question", func() {
			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--stemcell-output-directory", existingNonDirFile.Name(),
				"--stemcell-iaas", "aws",
				"--output-directory", validDirectory,
			})

			expectedOutput := fmt.Sprintf("--stemcell-output-directory %q is not a directory", existingNonDirFile.Name())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedOutput))
		})
	})

	When("pivnet-api-token is missing while no source is set", func() {
		It("returns an error", func() {
			tempDir, err := os.MkdirTemp("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command, []string{
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "mayhem-crew",
				"--product-version", `2.0.0`,
				"--output-directory", tempDir,
			})
			Expect(err).To(MatchError(`could not execute "download-product": could not parse download-product flags: missing required flag "--pivnet-api-token"`))
		})
	})

	When("both product-version and product-version-regex are set", func() {
		It("fails with an error saying that the user must pick one or the other", func() {
			tempDir, err := os.MkdirTemp("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--product-version-regex", ".*",
				"--output-directory", tempDir,
			})
			Expect(err).To(MatchError(ContainSubstring("cannot use both --product-version and --product-version-regex; please choose one or the other")))
		})
	})

	When("neither product-version nor product-version-regex are set", func() {
		It("fails with an error saying that the user must provide one or the other", func() {
			tempDir, err := os.MkdirTemp("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--output-directory", tempDir,
			})
			Expect(err).To(MatchError(ContainSubstring("no version information provided; please provide either --product-version or --product-version-regex")))
		})
	})

	When("the release specified is not available", func() {
		BeforeEach(func() {
			fakeProductDownloader.GetLatestProductFileReturns(nil, errors.New("some-error"))
		})

		It("returns an error", func() {
			tempDir, err := os.MkdirTemp("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = executeCommand(command, []string{
				"--pivnet-api-token", "token",
				"--file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--output-directory", tempDir,
			})
			Expect(err).To(MatchError(ContainSubstring("could not download product: some-error")))
		})
	})
})

func createProductPivotalFile(file *os.File) {
	var err error
	defer file.Close()

	z := zip.NewWriter(file)

	// https://github.com/pivotal-cf/om/issues/239
	// writing a "directory" as well, because some tiles seem to
	// have this as a separate file in the zip, which influences the regexp
	// needed to capture the metadata file
	_, err = z.Create("metadata/")
	Expect(err).ToNot(HaveOccurred())

	f, err := z.Create("metadata/fake-tile.yml")
	Expect(err).ToNot(HaveOccurred())

	_, err = f.Write([]byte(`{name: fake-tile, product_version: 1.2.3}`))
	Expect(err).ToNot(HaveOccurred())
	Expect(z.Close()).To(Succeed())
}
