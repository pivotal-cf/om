package metadata_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/om/configtemplate/metadata"
)

var _ = Describe("Pivnet Client", func() {
	When("too many files match the glob", func() {
		It("returns an error", func() {
			server := ghttp.NewServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ReleasesResponse{
						Releases: []pivnet.Release{
							{
								ID:      1,
								Version: "1.1.1",
							},
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/1/product_files"),
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
				),
			)

			provider := metadata.NewPivnetProvider(
				server.URL(),
				"some-token",
				"example-product",
				"1.1.1",
				"*.pivotal",
			)
			_, err := provider.MetadataBytes()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("the glob '*.pivotal' matches multiple files. Write your glob to match exactly one of the following:\n  something.pivotal\n  something-else.pivotal"))
		})
	})

	When("no versions match the given version", func() {
		It("returns an error", func() {
			server := ghttp.NewServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, pivnet.ReleasesResponse{
						Releases: []pivnet.Release{
							{
								ID:      1,
								Version: "2.2.2",
							},
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v2/products/example-product/releases/1/product_files"),
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
				),
			)

			provider := metadata.NewPivnetProvider(
				server.URL(),
				"some-token",
				"example-product",
				"1.1.1",
				"*.pivotal",
			)

			_, err := provider.MetadataBytes()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("no version matched for slug example-product, version 1.1.1 and glob *.pivotal"))
		})
	})
})
