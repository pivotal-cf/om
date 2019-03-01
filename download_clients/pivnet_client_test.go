package download_clients_test

import (
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
	log "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/pivotal-cf/om/download_clients"
	"github.com/pivotal-cf/om/download_clients/fakes"
)

var _ = Describe("PivnetClient", func() {
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

			fakePivnetFactory := func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", nil)
			versions, err := client.GetAllProductVersions("slug-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePivnetDownloader.ReleasesForProductSlugCallCount()).To(Equal(1))
			Expect(versions).To(ContainElement("1.0.0"))
			Expect(versions).To(ContainElement("2.0.0"))
		})
	})

	Context("GetLatestProductFile", func() {
		var (
			fakePivnetDownloader *fakes.PivnetDownloader
			fakePivnetFilter     *fakes.PivnetFilter
			logger               = &loggerfakes.FakeLogger{}
			fakePivnetFactory    func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFilter = &fakes.PivnetFilter{}
			fakePivnetFactory = func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})

		It("get the specific product file given a specific version and a slug", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)
			fakePivnetFilter.ProductFileKeysByGlobsReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			artifact, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
			Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
			Expect(fakePivnetFilter.ProductFileKeysByGlobsCallCount()).To(Equal(1))
			Expect(artifact.Name).To(Equal("someslug"))
		})

		It("returns an error if it could not find the release for the given slug and version pair", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease(""), errors.New("some error"))

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not fetch the release for someslug"))
		})

		It("returns an error if product files are not available for a slug", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{}, errors.New("some error"))

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not fetch the product files for someslug"))
		})

		It("returns an error could not understand the glob", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)
			fakePivnetFilter.ProductFileKeysByGlobsReturns([]pivnet.ProductFile{}, errors.New("couldn't understand blob"))

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not glob product files:"))
		})

		It("returns an error if there are no files that match the given glob", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)
			fakePivnetFilter.ProductFileKeysByGlobsReturns([]pivnet.ProductFile{}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*.zip' matches no file"))
		})

		It("returns an error if the glob matches multiple files", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)
			fakePivnetFilter.ProductFileKeysByGlobsReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*.zip' matches multiple files."))
		})
	})

	Context("DownloadProductToFile", func() {
		var (
			fakePivnetDownloader *fakes.PivnetDownloader
			fakePivnetFilter     *fakes.PivnetFilter
			logger               = &loggerfakes.FakeLogger{}
			fakePivnetFactory    func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFilter = &fakes.PivnetFilter{}
			fakePivnetFactory = func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})

		It("downloads a product file to given destination", func() {
			fakePivnetDownloader.DownloadProductFileReturns(nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			err := client.DownloadProductToFile(createFileArtifact(), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if the product file could not be downloaded", func() {
			fakePivnetDownloader.DownloadProductFileReturns(errors.New("download error"))

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			err := client.DownloadProductToFile(createFileArtifact(), nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not download product file"))
		})

	})

	Context("GetLatestStemcellForProduct", func() {
		var (
			fakePivnetDownloader     *fakes.PivnetDownloader
			fakePivnetFilter         *fakes.PivnetFilter
			logger                   = &loggerfakes.FakeLogger{}
			fakePivnetFactory        func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader
			errorTemplateForStemcell = "versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: %s"
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFilter = &fakes.PivnetFilter{}
			fakePivnetFactory = func(config pivnet.ClientConfig, logger log.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})
		It("downloads the stemcell", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			stemcell, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version).To(Equal("1.0"))
		})

		It("sets the minor version to 0 if only the major version is defined", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			stemcell, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version).To(Equal("1"))
		})

		It("downloads the latest major of the stemcell if multiple are available", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0", "someslug.stemcells"),
				createReleaseDependency(789, "0.10", "someslug.stemcells"),
				createReleaseDependency(789, "5.10", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			stemcell, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version).To(Equal("5.10"))
		})

		It("downloads the latest minor of a stemcell", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.3", "someslug.stemcells"),
				createReleaseDependency(789, "1.1", "someslug.stemcells"),
				createReleaseDependency(789, "1.2", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			stemcell, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version).To(Equal("1.3"))
		})

		It("returns an error if no stemcell is available for product", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{}, errors.New("stemcell not found"))

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not fetch stemcell dependency for"))
		})

		It("returns an error if the stemcell follows standard semver major.minor.patch format", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not sort stemcell dependency"))
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "1.0.0")))
		})

		It("returns an error if the major stemcell version contains an invalid character", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "abc1.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "abc1.0")))
		})

		It("returns an error if the minor stemcell version contains an invalid character", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0def", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(logger, nil, fakePivnetFactory, "", fakePivnetFilter)
			_, err := client.GetLatestStemcellForProduct(createFileArtifact(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "1.0def")))
		})
	})

})

func createRelease(version string) pivnet.Release {
	return pivnet.Release{
		Version: version,
		ID:      123,
	}
}

func createProductFile(slug string) pivnet.ProductFile {
	return pivnet.ProductFile{
		AWSObjectKey: slug,
		SHA256:       "somesha",
		ID:           456,
	}
}

func createFileArtifact() *download_clients.FileArtifact {
	return &download_clients.FileArtifact{}
}

func createReleaseDependency(id int, version string, slug string) pivnet.ReleaseDependency {
	return pivnet.ReleaseDependency{
		Release: pivnet.DependentRelease{
			ID:      id,
			Version: version,
			Product: pivnet.Product{
				Slug: slug,
			},
		},
	}
}
