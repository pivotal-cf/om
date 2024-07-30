package metadata_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"

	"github.com/pivotal-cf/om/configtemplate/metadata"
)

var _ = Describe("Pivnet Client", func() {
	When("too many files match the glob", func() {
		It("returns an error", func() {
			server := ghttp.NewServer()
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1",
				ghttp.RespondWith(http.StatusOK, `{"id":1}`),
			)
			server.RouteToHandler("POST", "/api/v2/products/example-product/releases/1/pivnet_resource_eula_acceptance",
				ghttp.RespondWith(http.StatusOK, `{}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1/file_groups",
				ghttp.RespondWith(http.StatusOK, `{}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases",
				ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ReleasesResponse{
					Releases: []pivnet.Release{
						{
							ID:      1,
							Version: "1.1.1",
						},
					},
				}),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1/product_files",
				ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ProductFilesResponse{
					ProductFiles: []pivnet.ProductFile{
						{
							ID:           1234,
							AWSObjectKey: "something.pivotal",
						},
						{
							ID:           2345,
							AWSObjectKey: "something-else.pivotal",
						},
					},
				}),
			)

			provider := metadata.NewPivnetProvider(server.URL(), "some-token", "example-product", "1.1.1", "*.pivotal", false)
			_, err := provider.MetadataBytes()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("for product version 1.1.1"))
		})
	})

	When("no versions match the given version", func() {
		It("returns an error", func() {
			server := ghttp.NewServer()
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1",
				ghttp.RespondWith(http.StatusOK, `{"id":1}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1/file_groups",
				ghttp.RespondWith(http.StatusOK, `{}`),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases",
				ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ReleasesResponse{
					Releases: []pivnet.Release{
						{
							ID:      1,
							Version: "2.2.2",
						},
						{
							ID:      1,
							Version: "3.3.3",
						},
					},
				}),
			)
			server.RouteToHandler("GET", "/api/v2/products/example-product/releases/1/product_files",
				ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ProductFilesResponse{
					ProductFiles: []pivnet.ProductFile{
						{
							ID:           1234,
							AWSObjectKey: "something.pivotal",
						},
						{
							ID:           2345,
							AWSObjectKey: "something-else.pivotal",
						},
					},
				}),
			)

			provider := metadata.NewPivnetProvider(server.URL(), "some-token", "example-product", "1.1.1", "*.pivotal", false)

			_, err := provider.MetadataBytes()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("could not fetch the release for example-product 1.1.1: release not found for version: '1.1.1'"))
		})
	})
})
