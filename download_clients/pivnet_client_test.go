package download_clients_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/onsi/gomega/ghttp"
	pivnetlog "github.com/pivotal-cf/go-pivnet/v6/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v6"

	"github.com/pivotal-cf/om/download_clients"
	"github.com/pivotal-cf/om/download_clients/fakes"
)

var _ = Describe("Grabbing Metadata", func() {
	When("a pivotal file", func() {
		It("returns the metadata", func() {
			stdout := log.New(GinkgoWriter, "", 0)
			stderr := log.New(GinkgoWriter, "", 0)
			modTime := time.Now()

			productFile := createPivotalFile("some.pivotal", "", "")
			contents, err := os.ReadFile(productFile)
			Expect(err).NotTo(HaveOccurred())

			server := ghttp.NewTLSServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.0.0"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases"),
					ghttp.RespondWith(http.StatusOK, `{
  "releases": [
    {
      "id": 24,
      "version": "1.0.0"
    }
  ]
}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases/24"),
					ghttp.RespondWith(http.StatusOK, `{"id":24}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/pivnet-product/releases/24/pivnet_resource_eula_acceptance"),
					ghttp.RespondWith(http.StatusOK, ""),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases/24/product_files"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{
  "product_files": [
  {
    "id": 1,
    "aws_object_key": "product.pivotal",
    "_links": {
      "download": {
        "href": "%s/api/v2/products/pivnet-product/releases/24/product_files/21/download"
      }
    }
  }
]
}`, server.URL())),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases/24/file_groups"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v2/products/pivnet-product/releases/24/product_files/21/download"),
					ghttp.RespondWith(http.StatusFound, `{}`, http.Header{"Location": {fmt.Sprintf("%s/api/v2/products/pivnet-product/releases/24/product_files/21/download", server.URL())}}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/api/v2/products/pivnet-product/releases/24/product_files/21/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/pivnet-product/releases/24/product_files/21/download"),
					func(w http.ResponseWriter, r *http.Request) {
						http.ServeContent(w, r, "download", modTime, bytes.NewReader(contents))
					},
				),
			)

			client := download_clients.NewPivnetClient(stdout, stderr, download_clients.DefaultPivnetFactory, "", true, server.URL())
			versions, err := client.GetAllProductVersions("pivnet-product")
			Expect(err).NotTo(HaveOccurred())
			Expect(versions).To(Equal([]string{"1.0.0"}))

			file, err := client.GetLatestProductFile("pivnet-product", versions[0], "*.pivotal")
			Expect(err).NotTo(HaveOccurred())

			metadata, err := file.ProductMetadata()
			Expect(err).NotTo(HaveOccurred())

			Expect(metadata.Name).To(Equal("example-product"))
			Expect(metadata.Version).To(Equal("1.0-build.0"))
		})
	})
})

var _ = Describe("PivnetClient", func() {
	var (
		stdout *log.Logger
		stderr *log.Logger
	)

	BeforeEach(func() {
		stdout = log.New(GinkgoWriter, "", 0)
		stderr = log.New(GinkgoWriter, "", 0)
	})

	Context("GetAllProductVersions", func() {
		var (
			fakePivnetDownloader *fakes.PivnetDownloader
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
		})

		It("gets a list of all releases", func() {
			fakePivnetDownloader.ReleasesForProductSlugReturns([]pivnet.Release{
				createRelease("1.0.0"),
				createRelease("2.0.0"),
			}, nil)

			fakePivnetFactory := func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
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
			fakePivnetFactory    func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFactory = func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})

		It("get the specific product file given a specific version and a slug", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug.zip"),
				createProductFile("anotherslug"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			artifact, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
			Expect(fakePivnetDownloader.ProductFilesForReleaseCallCount()).To(Equal(1))
			Expect(artifact.Name()).To(Equal("someslug.zip"))
		})

		It("returns an error if it could not find the release for the given slug and version pair", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease(""), errors.New("some error"))

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(MatchError(ContainSubstring("could not fetch the release for someslug")))
		})

		It("returns an error if it could not accept the EULA", func() {
			fakePivnetDownloader.AcceptEULAReturns(errors.New("some error"))

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(MatchError(ContainSubstring("could not accept EULA")))
		})

		It("returns an error if product files are not available for a slug", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{}, errors.New("some error"))

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(MatchError(ContainSubstring("could not fetch the product files for someslug")))
		})

		It("returns an error could not understand the glob", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*[")
			Expect(err).To(MatchError(ContainSubstring("could not glob product files:")))
		})

		It("returns an error if there are no files that match the given glob", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug"),
				createProductFile("anotherslug"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(MatchError(ContainSubstring("for product version 1.0.0: the glob '*.zip' matches no file")))
		})

		It("returns an error if the glob matches multiple files", func() {
			fakePivnetDownloader.ReleaseForVersionReturns(createRelease("1.0.0"), nil)
			fakePivnetDownloader.ProductFilesForReleaseReturns([]pivnet.ProductFile{
				createProductFile("someslug.zip"),
				createProductFile("anotherslug.zip"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestProductFile("someslug", "1.0.0", "*.zip")
			Expect(err).To(MatchError(ContainSubstring("the glob '*.zip' matches multiple files.")))
		})
	})

	Context("DownloadProductToFile", func() {
		var (
			fakePivnetDownloader *fakes.PivnetDownloader
			fakePivnetFactory    func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFactory = func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})

		It("downloads a product file to given destination", func() {
			fakePivnetDownloader.DownloadProductFileReturns(nil)
			tmpFile, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			err = client.DownloadProductToFile(createPivnetFileArtifact(), tmpFile)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if the product file could not be downloaded", func() {
			fakePivnetDownloader.DownloadProductFileReturns(errors.New("download error"))
			tmpFile, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			err = client.DownloadProductToFile(createPivnetFileArtifact(), tmpFile)
			Expect(err).To(MatchError(ContainSubstring("could not download product file")))
		})

	})

	Context("GetLatestStemcellForProduct", func() {
		var (
			fakePivnetDownloader     *fakes.PivnetDownloader
			fakePivnetFactory        func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader
			errorTemplateForStemcell = "versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: %s"
		)

		BeforeEach(func() {
			fakePivnetDownloader = &fakes.PivnetDownloader{}
			fakePivnetFactory = func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) download_clients.PivnetDownloader {
				return fakePivnetDownloader
			}
		})
		It("downloads the stemcell", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			stemcell, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version()).To(Equal("1.0"))
		})

		It("sets the minor version to 0 if only the major version is defined", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			stemcell, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version()).To(Equal("1"))
		})

		It("downloads the latest major of the stemcell if multiple are available", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0", "someslug.stemcells"),
				createReleaseDependency(789, "0.10", "someslug.stemcells"),
				createReleaseDependency(789, "5.10", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			stemcell, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version()).To(Equal("5.10"))
		})

		It("downloads the latest minor of a stemcell", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.3", "someslug.stemcells"),
				createReleaseDependency(789, "1.1", "someslug.stemcells"),
				createReleaseDependency(789, "1.2", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			stemcell, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version()).To(Equal("1.3"))
		})

		It("downloads the stemcell with the highest minor", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "97.190", "someslug.stemcells"),
				createReleaseDependency(789, "97.18", "someslug.stemcells"),
				createReleaseDependency(789, "97.9", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			stemcell, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).ToNot(BeNil())
			Expect(stemcell.Version()).To(Equal("97.190"))
		})

		It("returns an error if no stemcell is available for product", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{}, errors.New("stemcell not found"))

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).To(MatchError(ContainSubstring("could not fetch stemcell dependency for")))
		})

		It("returns an error if the stemcell follows standard semver major.minor.patch format", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).To(MatchError(ContainSubstring("could not sort stemcell dependency")))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "1.0.0"))))
		})

		It("returns an error if the major stemcell version contains an invalid character", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "abc1.0", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "abc1.0"))))
		})

		It("returns an error if the minor stemcell version contains an invalid character", func() {
			fakePivnetDownloader.ReleaseDependenciesReturns([]pivnet.ReleaseDependency{
				createReleaseDependency(789, "1.0def", "someslug.stemcells"),
			}, nil)

			client := download_clients.NewPivnetClient(stdout, stderr, fakePivnetFactory, "", true, "")
			_, err := client.GetLatestStemcellForProduct(createPivnetFileArtifact(), "", "")
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(errorTemplateForStemcell, "1.0def"))))
		})
	})

})

func createRelease(version string) pivnet.Release {
	return pivnet.Release{
		Version: version,
		ID:      123,
	}
}

func createProductFile(filename string) pivnet.ProductFile {
	return pivnet.ProductFile{
		AWSObjectKey: filename,
		SHA256:       "somesha",
		ID:           456,
	}
}

func createPivnetFileArtifact() download_clients.FileArtifacter {
	return &download_clients.PivnetFileArtifact{}
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
