package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	log "github.com/pivotal-cf/go-pivnet/logger"
)

var _ = FDescribe("PivnetClient", func() {
	Context("GetAllProductVersions", func() {
		var (
			fakePivnetDownloader *fakes.PivnetDownloader
			logger               = &loggerfakes.FakeLogger{}
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
		})

		It("gets a list of all releases", func() {
			fakePivnetDownloader.ReleasesForProductSlugReturns([]pivnet.Release{
				createRelease("1.0.0"),
				createRelease("2.0.0"),
			}, nil)

			fakePivnetFactory := func(config pivnet.ClientConfig, logger log.Logger) commands.PivnetDownloader {
				return fakePivnetDownloader
			}

			client := commands.NewPivnetClient(logger, nil, fakePivnetFactory, "")
			versions, err := client.GetAllProductVersions("slug-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePivnetDownloader.ReleasesForProductSlugCallCount()).To(Equal(1))
			Expect(versions).To(ContainElement("1.0.0"))
			Expect(versions).To(ContainElement("2.0.0"))
		})
	})

	Context("GetLatestProductFile", func() {
		It("gets the latest version for a given glob and slug", func() {

		})

		It("returns an error if it could not find the slug release", func() {

		})

		It("returns an error if product files are not available for a slug", func() {

		})

		It("returns an error could not understand the glob", func() {

		})

		It("returns an error if there are no files that match the given glob", func() {

		})

		It("returns an error if the glob matches multiple files", func() {

		})
	})

	Context("DownloadProductToFile", func() {
		It("downloads a product file to given destination", func() {

		})

		It("returns an error if the product file could not be downloaded", func() {

		})

	})

	Context("DownloadProductStemcell", func() {
		It("downloads the stemcell", func() {

		})

		It("returns an error if no stemcell is available for product", func() {

		})

		It("what is getLatestStemcell doing", func() {

		})
	})

})

func createRelease(version string) pivnet.Release{
	return pivnet.Release{
		Version: version,
	}
}