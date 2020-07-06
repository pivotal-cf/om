package commands_test

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/pivotal-cf/om/download_clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/download_clients/fakes"
	"github.com/pivotal-cf/om/validator"
)

var _ = Describe("DownloadProduct", func() {
	var (
		command               *commands.DownloadProduct
		environFunc           func() []string
		err                   error
		fakeProductDownloader *fakes.ProductDownloader
		buffer                *gbytes.Buffer
	)

	BeforeEach(func() {
		fakeProductDownloader = &fakes.ProductDownloader{}
		fakeProductDownloader.GetAllProductVersionsReturns([]string{"2.0.0"}, nil)
		environFunc = func() []string { return nil }
	})

	JustBeforeEach(func() {
		download_clients.NewPivnetClient = func(stdout *log.Logger, stderr *log.Logger, factory download_clients.PivnetFactory, token string, skipSSL bool, pivnetHost string) download_clients.ProductDownloader {
			return fakeProductDownloader
		}
		buffer = gbytes.NewBuffer()
		command = commands.NewDownloadProduct(
			environFunc,
			log.New(buffer, "", 0),
			log.New(buffer, "", 0),
			buffer,
		)
	})

	When("the flags are set correctly", func() {
		When("it can connect to the source", func() {
			BeforeEach(func() {
				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")

				fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
			})

			It("downloads a product from the downloader", func() {
				tempDir, err := ioutil.TempDir("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				}

				err = command.Execute(commandArgs)
				Expect(err).ToNot(HaveOccurred())
			})

			It("supports the pivnet-file-glob alias for file-glob", func() {
				tempDir, err := ioutil.TempDir("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				}

				err = command.Execute(commandArgs)
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
				tempDir, err := ioutil.TempDir("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())

				commandArgs := []string{
					"--pivnet-api-token", "token",
					"--file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version-regex", `2\..\..*`,
					"--output-directory", tempDir,
				}

				err = command.Execute(commandArgs)
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version-regex", `2\..\..*`,
						"--output-directory", tempDir,
					}

					err = command.Execute(commandArgs)
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version-regex", `2\..\..*`,
						"--output-directory", tempDir,
					}

					err = command.Execute(commandArgs)
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
						return ioutil.WriteFile(file.Name(), []byte("contents"), 0777)
					}
				})

				It("downloads a product from the downloader", func() {
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					}

					err = command.Execute(commandArgs)
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
						return ioutil.WriteFile(file.Name(), []byte("contents"), 0777)
					}
				})

				It("errors and removes the file from the file system", func() {
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					commandArgs := []string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					}

					err = command.Execute(commandArgs)
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
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

					err = command.Execute(commandArgs)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeProductDownloader.GetLatestStemcellForProductCallCount()).To(Equal(1))
					Expect(fakeProductDownloader.GetLatestProductFileCallCount()).To(Equal(2))
					Expect(fakeProductDownloader.DownloadProductToFileCallCount()).To(Equal(2))
					Expect(fakeProductDownloader.GetAllProductVersionsCallCount()).To(Equal(1))

					fa, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(1)
					Expect(fa.Name()).To(Equal("stemcell.tgz"))
					Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "stemcell.tgz.partial")))

					fileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(fileName)
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
					fileContent, err = ioutil.ReadFile(fileName)
					Expect(err).ToNot(HaveOccurred())
					Expect(fileName).To(BeAnExistingFile())
					Expect(string(fileContent)).To(MatchJSON(`
							{
								"product": "fake-tile",
								"stemcell": "97.190"
							}`))
				})

				When("the --stemcell-output-dir flag is passed", func() {
					var (
						commandArgs       []string
						productOutputDir  string
						stemcellOutputDir string
					)

					tempFile := func(dir, pattern string) string {
						file, err := ioutil.TempFile(dir, pattern)
						Expect(err).ToNot(HaveOccurred())
						return file.Name()
					}

					BeforeEach(func() {
						productOutputDir, err = ioutil.TempDir("", "om-tests-output-dir-")
						Expect(err).ToNot(HaveOccurred())

						stemcellOutputDir, err = ioutil.TempDir("", "om-tests-stemcell-output-dir-")
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
						err = command.Execute(commandArgs)
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

							err = command.Execute(commandArgs)
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

							err = command.Execute(commandArgs)
							Expect(err).ToNot(HaveOccurred())

							downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
							downloadedStemcellFilePath := path.Join(stemcellOutputDir, "stemcell.tgz")
							Expect(downloadedFilePath).To(BeAnExistingFile())
							Expect(downloadedStemcellFilePath).To(BeAnExistingFile())

							Expect(alreadyDownloadedProduct).ToNot(BeAnExistingFile())
							Expect(alreadyDownloadedLightStemcell).ToNot(BeAnExistingFile())
							Expect(unknownFileWeDontOwn).To(BeAnExistingFile())
						})

						When("the product has already been downloaded and cached", func() {
							It("only deletes previous versions of the product", func() {
								previousDownloadedProduct := tempFile(productOutputDir, "cf-2.0-build.*.pivotal")
								downloadedFilePath := path.Join(productOutputDir, "cf-2.0-build.1.pivotal")
								downloadedFile, err := os.Create(downloadedFilePath)
								Expect(err).ToNot(HaveOccurred())
								createProductPivotalFile(downloadedFile)

								err = command.Execute(commandArgs)
								Expect(err).ToNot(HaveOccurred())

								downloadedStemcellFilePath := path.Join(stemcellOutputDir, "stemcell.tgz")
								Expect(downloadedFilePath).To(BeAnExistingFile())
								Expect(downloadedStemcellFilePath).To(BeAnExistingFile())

								Expect(previousDownloadedProduct).ToNot(BeAnExistingFile())
							})
						})
					})
				})

				When("the --stemcell-version flag is passed", func() {
					It("downloads the specified stemcell at that version", func() {
						tempDir, err := ioutil.TempDir("", "om-tests-")
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

						err = command.Execute(commandArgs)
						Expect(err).ToNot(HaveOccurred())

						fa, pf := fakeProductDownloader.DownloadProductToFileArgsForCall(1)
						Expect(fa.Name()).To(Equal("stemcell.tgz"))
						Expect(pf.Name()).To(Equal(filepath.Join(tempDir, "[stemcells-ubuntu-xenial,100.00]stemcell.tgz.partial")))

						fileName := path.Join(tempDir, "download-file.json")
						fileContent, err := ioutil.ReadFile(fileName)
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
						fileContent, err = ioutil.ReadFile(fileName)
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.tgz",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
						"--stemcell-iaas", "google",
					})

					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
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

					tempDir, err := ioutil.TempDir("", "om-tests-")
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

					err = command.Execute(commandArgs)
					Expect(err).To(MatchError(ContainSubstring("No stemcell identified for IaaS \"unknown-iaas\" on Pivotal Network. Correct the `stemcell-iaas` option to match the IaaS portion of the stemcell filename, or remove the option")))
				})
			})

			When("the --stemcell-heavy flag is not provided", func() {
				It("downloads the corresponding light or heavy stemcell", func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
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

					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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

					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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

					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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

		When("the file is already downloaded", func() {
			var tempDir string

			BeforeEach(func() {
				tempDir, err = ioutil.TempDir("", "om-tests-")
				Expect(err).ToNot(HaveOccurred())
			})

			setupForProductAPI := func(shaSum string) {
				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
				fa.SHA256Returns(shaSum)
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
			}

			createFilePath := func() string {
				filePath := path.Join(tempDir, "cf-2.0-build.1.pivotal")
				file, err := os.Create(filePath)
				Expect(err).ToNot(HaveOccurred())
				createProductPivotalFile(file)
				return filePath
			}

			When("a sha sum is provided by the downloader", func() {
				BeforeEach(func() {
					filePath := createFilePath()

					validator := validator.NewSHA256Calculator()
					sum, err := validator.Checksum(filePath)
					Expect(err).ToNot(HaveOccurred())
					setupForProductAPI(sum)

					os.Setenv("CACHE_CLEANUP", "I acknowledge this will delete files in the output directories")
				})

				AfterEach(func() {
					os.Unsetenv("CACHE_CLEANUP")
				})

				It("does not download the file again, even with cache cleanup", func() {
					err = command.Execute([]string{
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

				It("does not panic when downloading the stemcell if file already downloaded", func() {
					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--stemcell-iaas", "google",
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("no sha sum is provided by downloader", func() {
				It("does not re-download the product", func() {
					createFilePath()
					setupForProductAPI("")

					err = command.Execute([]string{
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
			})

			When("the sha is invalid", func() {
				It("downloads it, again", func() {
					createFilePath()
					setupForProductAPI("20a9668171397bf4ea9487835e28e9ca090f3b04d1d0461f8d3b752a3e0daf30")

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "elastic-runtime",
						"--product-version", "2.0.0",
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		When("the --config flag is passed", func() {
			BeforeEach(func() {
				fa := &fakes.FileArtifacter{}
				fa.NameReturns("/some-account/some-bucket/cf-2.0-build.1.pivotal")
				fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
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

		Describe("managing and reporting the filename written to the filesystem", func() {
			When("S3 configuration is provided and source is not set", func() {
				BeforeEach(func() {
					fa := &fakes.FileArtifacter{}
					fa.NameReturns("/some-account/some-bucket/my-great-product.pivotal")
					fakeProductDownloader.GetLatestProductFileReturns(fa, nil)
				})

				It("prefixes the filename with a bracketed slug and version", func() {
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
						"--s3-bucket", "there once was a man from a",
					})
					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
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
					tempDir, err := ioutil.TempDir("", "om-tests-")
					Expect(err).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--pivnet-api-token", "token",
						"--file-glob", "*.pivotal",
						"--pivnet-product-slug", "mayhem-crew",
						"--product-version", `2.0.0`,
						"--output-directory", tempDir,
					})
					Expect(err).ToNot(HaveOccurred())

					downloadReportFileName := path.Join(tempDir, "download-file.json")
					fileContent, err := ioutil.ReadFile(downloadReportFileName)
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
			tempDir, err := ioutil.TempDir("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
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
			validDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors, printing both --output-dir and the filepath in question", func() {
			err := command.Execute([]string{
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
			err := command.Execute([]string{
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
			existingNonDirFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			validDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("errors, printing both --output-directory and the filepath in question", func() {
			err = command.Execute([]string{
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
			err = command.Execute([]string{
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

	When("an unknown flag is provided", func() {
		It("returns an error", func() {
			err = command.Execute([]string{"--badflag"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not parse download-product flags: flag provided but not defined: -badflag"))
		})
	})

	When("a required flag is not provided", func() {
		It("returns an error", func() {
			err = command.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not parse download-product flags: missing required flag \"--output-directory\""))
		})
	})

	When("pivnet-api-token is missing while no source is set", func() {
		It("returns an error", func() {
			tempDir, err := ioutil.TempDir("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
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
			tempDir, err := ioutil.TempDir("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
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
			tempDir, err := ioutil.TempDir("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
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
			fakeProductDownloader.GetLatestProductFileReturns(nil, fmt.Errorf("some-error"))
		})

		It("returns an error", func() {
			tempDir, err := ioutil.TempDir("", "om-tests-")
			Expect(err).ToNot(HaveOccurred())

			err = command.Execute([]string{
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
