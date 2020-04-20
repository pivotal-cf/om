package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("ConfigureOpsmanService", func() {
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

	Describe("UpdatePivnetToken", func() {
		It("updates the pivnet token associated with the ops manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/settings/pivotal_network_settings"),
					ghttp.RespondWith(http.StatusOK, `{
					  "success": true
					}`),
					ghttp.VerifyJSON("{ \"pivotal_network_settings\": { \"api_token\": \"some-api-token\" }}"),
				),
			)

			err := service.UpdatePivnetToken("some-api-token")
			Expect(err).ToNot(HaveOccurred())
		})

		When("the api returns an error", func() {
			It("returns the error to the user", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/settings/pivotal_network_settings"),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					),
				)

				err := service.UpdatePivnetToken("some-api-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("500 Internal Server Error"))
			})
		})
	})

	Describe("UpdateSSLCertificate", func() {
		It("updates the ssl certificate in ops manager settings", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/settings/ssl_certificate"),
					ghttp.RespondWith(http.StatusOK, `{
					  "ssl_certificate": {
						"certificate": "some-cert",
						"private_key": "some-key"
					  }
					}`),
					ghttp.VerifyJSON(`{"certificate": "some-cert","private_key": "some-key"}`),
				),
			)

			certInput := api.SSLCertificateInput{
				CertPem:       "some-cert",
				PrivateKeyPem: "some-key",
			}
			err := service.UpdateSSLCertificate(certInput)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the api returns an error", func() {
			It("returns the error to the user", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/settings/ssl_certificate"),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					),
				)

				certInput := api.SSLCertificateInput{
					CertPem:       "some-cert",
					PrivateKeyPem: "some-key",
				}
				err := service.UpdateSSLCertificate(certInput)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("500 Internal Server Error"))
			})
		})
	})

	Describe("GetSSLCertificate", func() {
		It("gets the ssl certificate from ops manager settings", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/settings/ssl_certificate"),
					ghttp.RespondWith(http.StatusOK, `{
					  "ssl_certificate": {
					    "certificate": "some-cert"
					  }
					}`),
				),
			)

			output, err := service.GetSSLCertificate()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Certificate.Certificate).To(Equal("some-cert"))
		})

		When("the api returns an error", func() {
			It("returns the error to the user", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/settings/ssl_certificate"),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					),
				)

				_, err := service.GetSSLCertificate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("500 Internal Server Error"))
			})
		})
	})

	Describe("DeleteSSLCertificate", func() {
		It("deletes the ssl certificate in ops manager settings", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/settings/ssl_certificate"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.DeleteSSLCertificate()
			Expect(err).ToNot(HaveOccurred())
		})

		When("the api returns an error", func() {
			It("returns the error to the user", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/settings/ssl_certificate"),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					),
				)

				err := service.DeleteSSLCertificate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("500 Internal Server Error"))
			})
		})
	})
})
