package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
	"os"
	"path"
)

var _ = FDescribe("DownloadProduct", func() {
	var (
		command              commands.DownloadProduct
		logger               *loggerfakes.FakeLogger
		fakeFactory          *fakes.PivnetClientFactory
		fakePivnetDownloader *fakes.PivnetDownloader
		fakeWriter           *gbytes.Buffer
		tempDir              string
		err                  error
	)

	BeforeEach(func() {
		logger = &loggerfakes.FakeLogger{}
		fakePivnetDownloader = &fakes.PivnetDownloader{}
		fakeFactory = &fakes.PivnetClientFactory{}
		fakeFactory.NewClientReturns(fakePivnetDownloader)
		fakeWriter = gbytes.NewBuffer()
	})

	JustBeforeEach(func() {
		command = commands.NewDownloadProduct(logger, fakeWriter, fakeFactory)
	})

	Context("given all the flags are set correctly", func() {
		BeforeEach(func() {
			fakePivnetDownloader.ReleaseForVersionReturns(pivnet.Release{
				ID: 12345,
			}, nil)

			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				{
					ID:           54321,
					AWSObjectKey: "/some-account/some-bucket/cf-2.0-build.1.pivotal",
					Name:         "cf-2.0-build.1.pivotal",
				},
			}, nil)

			tempDir, err = ioutil.TempDir("", "om-tests-")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err = os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("downloads a product from Pivotal Network", func() {
			err := command.Execute([]string{
				"--pivnet-api-token", "token",
				"--pivnet-file-glob", "*.pivotal",
				"--pivnet-product-slug", "elastic-runtime",
				"--product-version", "2.0.0",
				"--output-directory", tempDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFactory.NewClientCallCount()).To(Equal(1))
			Expect(fakePivnetDownloader.ReleaseForVersionCallCount()).To(Equal(1))
			Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
			Expect(fakePivnetDownloader.DownloadProductFileCallCount()).To(Equal(1))

			slug, version := fakePivnetDownloader.ReleaseForVersionArgsForCall(0)
			Expect(slug).To(Equal("elastic-runtime"))
			Expect(version).To(Equal("2.0.0"))

			slug, releaseID := fakePivnetDownloader.ProductFilesForReleaseArgsForCall(0)
			Expect(slug).To(Equal("elastic-runtime"))
			Expect(releaseID).To(Equal(12345))

			file, slug, releaseID, productFileID, _ := fakePivnetDownloader.DownloadProductFileArgsForCall(0)
			Expect(file.Name()).To(Equal(path.Join(tempDir, "cf-2.0-build.1.pivotal")))
			Expect(slug).To(Equal("elastic-runtime"))
			Expect(releaseID).To(Equal(12345))
			Expect(productFileID).To(Equal(54321))
		})

		Context("when the globs returns multiple files", func() {
			BeforeEach(func() {
				fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
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
				err := command.Execute([]string{
					"--pivnet-api-token", "token",
					"--pivnet-file-glob", "*.pivotal",
					"--pivnet-product-slug", "elastic-runtime",
					"--product-version", "2.0.0",
					"--output-directory", tempDir,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`the glob '*.pivotal' matches multiple files. Write your glob to match exactly one of the following: [cf-2.0-build.1.pivotal srt-2.0-build.1.pivotal]`))
			})
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewDownloadProduct(logger, fakeWriter, fakeFactory)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse download-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when a required flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewDownloadProduct(logger, fakeWriter, fakeFactory)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse download-product flags: missing required flag \"--pivnet-api-token\""))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDownloadProduct(logger, fakeWriter, fakeFactory)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
				ShortDescription: "downloads a specified product file from Pivotal Network",
				Flags:            command.Options,
			}))
		})
	})
})
