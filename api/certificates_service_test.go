package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Certificates", func() {
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

	Describe("GenerateCertificate", func() {
		It("returns a cert and key", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/certificates/generate"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
						"domains": [
							"*.example.com",
							"*.example.org"
						]
					}`),
					ghttp.RespondWith(http.StatusOK, `{
						"certificate": "some-certificate",
						"key": "some-key"
					}`),
				),
			)

			output, err := service.GenerateCertificate(api.DomainsInput{
				Domains: []string{"*.example.com", "*.example.org"},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchJSON(`{
				"certificate": "some-certificate",
				"key": "some-key"
			}`))
		})

		When("the client cannot make the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GenerateCertificate(api.DomainsInput{Domains: []string{"some-domains"}})
				Expect(err).To(MatchError(ContainSubstring("could not send api request to POST /api/v0/certificates/generate")))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificates/generate"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.GenerateCertificate(api.DomainsInput{Domains: []string{"some-domains"}})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
