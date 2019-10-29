package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Security", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{client.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	It("gets the root CA cert", func() {
		client.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/security/root_ca_certificate"),
				ghttp.RespondWith(http.StatusOK, `{"root_ca_certificate_pem": "some-response-cert"}`),
			),
		)

		output, err := service.GetSecurityRootCACertificate()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal("some-response-cert"))
	})

	It("returns error if request fails to submit", func() {
		client.Close()

		_, err := service.GetSecurityRootCACertificate()
		Expect(err).To(MatchError(ContainSubstring("failed to submit request")))
	})

	It("returns error when response contains non-200 status code", func() {
		client.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/security/root_ca_certificate"),
				ghttp.RespondWith(http.StatusTeapot, `{}`),
			),
		)

		_, err := service.GetSecurityRootCACertificate()
		Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
	})

	It("returns error if response fails to unmarshal", func() {
		client.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/security/root_ca_certificate"),
				ghttp.RespondWith(http.StatusOK, `%%%`),
			),
		)

		_, err := service.GetSecurityRootCACertificate()
		Expect(err).To(MatchError(ContainSubstring("failed to unmarshal response: invalid character")))
	})
})
