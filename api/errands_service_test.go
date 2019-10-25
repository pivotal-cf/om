package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("ErrandsService", func() {
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

	Describe("UpdateStagedProductErrands", func() {
		It("sets state for a product's errands", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-id/errands"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
						"errands": [{
							"name": "some-errand",
							"post_deploy": "when-changed",
							"pre_delete": false
						}]
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.UpdateStagedProductErrands("some-product-id", "some-errand", "when-changed", false)
			Expect(err).ToNot(HaveOccurred())
		})

		When("ops manager returns a not-OK response code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-id/errands"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.UpdateStagedProductErrands("some-product-id", "some-errand", "when-changed", "false")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the product ID cannot be URL encoded", func() {
			It("returns an error", func() {
				err := service.UpdateStagedProductErrands("%%%", "some-errand", "true", "false")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				err := service.UpdateStagedProductErrands("some-product-id", "some-errand", "true", "false")
				Expect(err).To(MatchError(ContainSubstring("failed to set errand state: could not send api request to PUT /api/v0/staged/products/some-product-id/errands")))
			})
		})
	})

	Describe("ListStagedProductErrands", func() {
		It("lists errands for a product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-id/errands"),
					ghttp.RespondWith(http.StatusOK, `{
					"errands": [
							{"post_deploy":"true","name":"first-errand"},
							{"post_deploy":"false","name":"second-errand"},
							{"pre_delete":"true","name":"third-errand"}
						]
					}`),
				),
			)

			output, err := service.ListStagedProductErrands("some-product-id")
			Expect(err).ToNot(HaveOccurred())

			Expect(output.Errands).To(ConsistOf([]api.Errand{
				{Name: "first-errand", PostDeploy: "true"},
				{Name: "second-errand", PostDeploy: "false"},
				{Name: "third-errand", PreDelete: "true"},
			}))
		})

		When("the product ID cannot be URL encoded", func() {
			It("returns an error", func() {
				_, err := service.ListStagedProductErrands("%%%")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListStagedProductErrands("some-product-id")
				Expect(err).To(MatchError(ContainSubstring("failed to list errands: could not send api request to GET /api/v0/staged/products/some-product-id/errands")))
			})
		})

		When("the response body cannot be parsed", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-id/errands"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStagedProductErrands("some-product-id")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})

		When("the http call returns an error status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/future-moon-and-assimilation/errands"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListStagedProductErrands("future-moon-and-assimilation")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
