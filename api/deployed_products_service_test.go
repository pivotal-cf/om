package api_test

import (
	"github.com/onsi/gomega/ghttp"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("DeployedProducts", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()
		service = api.New(api.ApiInput{
			Client: httpClient{serverURI: client.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("GetDeployedProductManifest", func() {
		It("returns a manifest of a product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/manifest"),
					ghttp.RespondWith(http.StatusOK, `{
						"key-1": {
							"key-2": "value-1"
						},
						"key-3": "value-2",
						"key-4": 2147483648
					}`),
				),
			)

			manifest, err := service.GetDeployedProductManifest("some-product-guid")
			Expect(err).ToNot(HaveOccurred())
			Expect(manifest).To(MatchYAML(`---
key-1:
  key-2: value-1
key-3: value-2
key-4: 2147483648
`))
		})

		When("the request object is invalid", func() {
			It("returns an error", func() {
				_, err := service.GetDeployedProductManifest("invalid-guid-%%%")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		When("the client request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetDeployedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not make api request to staged products manifest endpoint: could not send api request to GET /api/v0/deployed/products/some-product-guid/manifest")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/manifest"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.GetDeployedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the returned JSON is invalid", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/manifest"),
						ghttp.RespondWith(http.StatusOK, `%%%`),
					),
				)

				_, err := service.GetDeployedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not parse json")))
			})
		})
	})

	Describe("List", func() {
		It("retrieves a list of deployed products from the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"guid":"some-product-guid",
						"type":"some-type"
					}, {
						"guid":"some-other-product-guid",
						"type":"some-other-type"
					}]`),
				),
			)

			output, err := service.ListDeployedProducts()
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(Equal([]api.DeployedProductOutput{{
				GUID: "some-product-guid",
				Type: "some-type",
			}, {
				GUID: "some-other-product-guid",
				Type: "some-other-type",
			}}))
		})

		When("the request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListDeployedProducts()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to deployed products endpoint: could not send api request to GET /api/v0/deployed/products")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListDeployedProducts()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the server returns invalid JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, "%%"),
					),
				)
				_, err := service.ListDeployedProducts()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal deployed products response:")))
			})
		})
	})
})
