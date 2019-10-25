package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
	"strings"
)

var _ = Describe("Available Products", func() {
	var (
		progressClient *ghttp.Server
		client         *ghttp.Server
		service        api.Api
	)

	BeforeEach(func() {
		progressClient = ghttp.NewServer()
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client:         httpClient{client.URL()},
			ProgressClient: httpClient{progressClient.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("UploadAvailableProduct", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			progressClient.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/available_products"),
					ghttp.VerifyContentType("some content-type"),
					ghttp.VerifyBody([]byte("some content")),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Expect(req.ContentLength).To(Equal(int64(12)))

						_, err := w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)

			output, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
				ContentLength:   12,
				Product:         strings.NewReader("some content"),
				ContentType:     "some content-type",
				PollingInterval: 1,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.UploadAvailableProductOutput{}))
		})

		When("an error occurs", func() {
			When("the client errors performing the request", func() {
				It("returns an error", func() {
					progressClient.Close()

					_, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to available_products endpoint")))
				})
			})

			When("the api returns a non-200 status code", func() {
				It("returns an error", func() {
					progressClient.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/available_products"),
							ghttp.RespondWith(http.StatusTeapot, `{}`),
						),
					)

					_, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("ListAvailableProducts", func() {
		It("lists available products", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/available_products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"name": "available-product",
						"product_version": "available-version"
					}]`),
				),
			)

			output, err := service.ListAvailableProducts()
			Expect(err).ToNot(HaveOccurred())

			Expect(output.ProductsList).To(ConsistOf([]api.ProductInfo{{
				Name:    "available-product",
				Version: "available-version",
			}}))
		})

		When("the client can't connect to the client", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListAvailableProducts()
				Expect(err).To(MatchError(ContainSubstring("could not make api request")))
			})
		})

		When("the client won't fetch available products", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.ListAvailableProducts()
				Expect(err).To(MatchError(ContainSubstring("request failed")))
			})
		})

		When("the response is not JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListAvailableProducts()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
			})
		})
	})

	Describe("DeleteAvailableProducts", func() {
		It("deletes a named product / version", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/available_products", "product_name=some-product&version=1.2.3-build.4"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
				ProductName:    "some-product",
				ProductVersion: "1.2.3-build.4",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("the ShouldDeleteAllProducts flag is provided", func() {
			It("does not provide a product query to DELETE", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/available_products", ""),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
					ShouldDeleteAllProducts: true,
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("a non-200 status code is returned", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
					ProductName:    "some-product",
					ProductVersion: "1.2.3-build.4",
				})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
