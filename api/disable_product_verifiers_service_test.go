package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
	"net/url"
)

type httpClient struct {
	serverURI string
}

func (h httpClient) Do(req *http.Request) (*http.Response, error) {
	uri, err := url.Parse(h.serverURI)
	Expect(err).ToNot(HaveOccurred())

	req.URL.Host = uri.Host
	req.URL.Scheme = uri.Scheme
	return http.DefaultClient.Do(req)
}

var _ = Describe("DisableProductVerifiersService", func() {
	var (
		server  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{server.URL()},
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ListProductVerifiers", func() {
		It("lists available verifiers for a product", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/cf-guid/verifiers/install_time"),
					ghttp.RespondWith(http.StatusOK, `{
						"verifiers": [
							{
								"type": "some-verifier-type",
								"enabled": true
							},
							{
								"type": "another-verifier-type",
								"enabled": false
							}
						]
					}`),
				),
			)

			verifiers, guid, err := service.ListProductVerifiers("cf")
			Expect(err).ToNot(HaveOccurred())

			Expect(verifiers).To(Equal([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: true,
				},
				{
					Type:    "another-verifier-type",
					Enabled: false,
				},
			}))

			Expect(guid).To(Equal("cf-guid"))
		})

		Context("failure cases", func() {
			It("returns an error when the staged/products endpoint returns not 200-OK", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusInternalServerError, `[]`),
				)

				_, _, err := service.ListProductVerifiers("cf")
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `[{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"}]`),
					ghttp.RespondWith(http.StatusInternalServerError, `[]`),
				)

				_, _, err := service.ListProductVerifiers("cf")
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error when the http request could not be made", func() {
				server.Close()

				_, _, err := service.ListProductVerifiers("cf")
				Expect(err).To(MatchError(ContainSubstring("could not make request")))
			})

			It("returns an error when the list_product_verifiers response is not JSON", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `[{"installation_name":"cf","guid":"cf-guid","type":"cf","product_version":"1.10.0-build.177"}]`),
					ghttp.RespondWith(http.StatusOK, `invalid json`),
				)

				_, _, err := service.ListProductVerifiers("cf")
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal list_product_verifiers response")))
			})
		})
	})

	Describe("DisableProductVerifiers", func() {
		It("disables a list of product verifiers", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/cf-guid/verifiers/install_time/some-verifier-type"),
					ghttp.RespondWith(http.StatusOK, `{"type":"some-verifier-type", "enabled":false}`),
					ghttp.VerifyJSON(`{"enabled": false}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/cf-guid/verifiers/install_time/another-verifier-type"),
					ghttp.RespondWith(http.StatusOK, `{"type":"another-verifier-type", "enabled":false}`),
					ghttp.VerifyJSON(`{"enabled": false}`),
				),
			)

			err := service.DisableProductVerifiers([]string{"some-verifier-type", "another-verifier-type"}, "cf-guid")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the endpoint returns a non-200-OK status code", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusInternalServerError, `[]`),
				)

				err := service.DisableProductVerifiers([]string{"some-verifier-type"}, "cf-guid")
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error when the http request could not be made", func() {
				server.Close()

				err := service.DisableProductVerifiers([]string{"some-verifier-type"}, "cf-guid")
				Expect(err).To(MatchError(ContainSubstring("could not make api request to disable_product_verifiers endpoint")))
			})
		})
	})
})
