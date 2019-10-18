package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("CertificateAuthorities", func() {
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

	Describe("ListCertificateAuthorities", func() {
		It("returns a slice of CAs", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/certificate_authorities"),
					ghttp.RespondWith(http.StatusOK, `{
						"certificate_authorities": [{
							"guid": "some-guid",
							"issuer": "some-issuer",
							"created_on": "2017-01-09",
							"expires_on": "2021-01-09",
							"active": true,
							"cert_pem": "some-cert-pem"
						}, {
							"guid": "some-guid-2",
							"issuer": "another-issuer",
							"created_on": "2017-09-09",
							"expires_on": "2021-10-02",
							"active": false,
							"cert_pem": "some-other-cert-pem"
						}]
					}`),
				),
			)

			output, err := service.ListCertificateAuthorities()
			Expect(err).NotTo(HaveOccurred())

			Expect(output.CAs).To(ConsistOf([]api.CA{{
				GUID:      "some-guid",
				Issuer:    "some-issuer",
				CreatedOn: "2017-01-09",
				ExpiresOn: "2021-01-09",
				Active:    true,
				CertPEM:   "some-cert-pem",
			}, {
				GUID:      "some-guid-2",
				Issuer:    "another-issuer",
				CreatedOn: "2017-09-09",
				ExpiresOn: "2021-10-02",
				Active:    false,
				CertPEM:   "some-other-cert-pem",
			}}))
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListCertificateAuthorities()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to GET /api/v0/certificate_authorities"))
			})
		})

		When("the response body is invalid json", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/certificate_authorities"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListCertificateAuthorities()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/certificate_authorities"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListCertificateAuthorities()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

	})

	Describe("GenerateCertificateAuthority", func() {
		It("generates a certificate authority", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/generate"),
					ghttp.RespondWith(http.StatusOK, `{
						"guid": "some-guid",
						"issuer": "some-issuer",
						"created_on": "2017-01-09",
						"expires_on": "2021-01-09",
						"active": true,
						"cert_pem": "some-cert-pem"
					}`),
				),
			)

			ca, err := service.GenerateCertificateAuthority()
			Expect(err).NotTo(HaveOccurred())

			Expect(ca).To(Equal(api.CA{
				GUID:      "some-guid",
				Issuer:    "some-issuer",
				CreatedOn: "2017-01-09",
				ExpiresOn: "2021-01-09",
				Active:    true,
				CertPEM:   "some-cert-pem",
			}))
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GenerateCertificateAuthority()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to POST /api/v0/certificate_authorities/generate"))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/generate"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.GenerateCertificateAuthority()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the response body is not valid json", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/generate"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GenerateCertificateAuthority()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})

	Describe("RegenerateCertificates", func() {
		It("regenerates certificate authority", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/active/regenerate"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.RegenerateCertificates()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				err := service.RegenerateCertificates()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to POST /api/v0/certificate_authorities/active/regenerate"))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/active/regenerate"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err := service.RegenerateCertificates()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("CreateCertificateAuthority", func() {
		var (
			certPem    string
			privateKey string
		)

		BeforeEach(func() {
			certPem = "some-cert"
			privateKey = "some-key"
		})

		It("creates a certificate authority", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{"cert_pem": "some-cert", "private_key_pem": "some-key"}`),
					ghttp.RespondWith(http.StatusOK, `{
						"guid": "some-guid",
						"issuer": "some-issuer",
						"created_on": "2017-01-09",
						"expires_on": "2021-01-09",
						"active": true,
						"cert_pem": "some-cert"
					}`),
				),
			)

			ca, err := service.CreateCertificateAuthority(api.CertificateAuthorityInput{
				CertPem:       certPem,
				PrivateKeyPem: privateKey,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(ca).To(Equal(api.CA{
				GUID:      "some-guid",
				Issuer:    "some-issuer",
				CreatedOn: "2017-01-09",
				ExpiresOn: "2021-01-09",
				Active:    true,
				CertPEM:   "some-cert",
			}))
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.CreateCertificateAuthority(api.CertificateAuthorityInput{
					CertPem:       certPem,
					PrivateKeyPem: privateKey,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to POST /api/v0/certificate_authorities"))
			})
		})

		When("the response body is invalid json", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.CreateCertificateAuthority(api.CertificateAuthorityInput{
					CertPem:       certPem,
					PrivateKeyPem: privateKey,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})

		When("it returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)
				_, err := service.CreateCertificateAuthority(api.CertificateAuthorityInput{
					CertPem:       certPem,
					PrivateKeyPem: privateKey,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request failed: unexpected response"))
			})
		})
	})

	Describe("ActivateCertificateAuthority", func() {
		It("activates a certificate authority", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/some-certificate-authority-guid/activate"),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
				GUID: "some-certificate-authority-guid",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				err := service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
					GUID: "some-certificate-authority-guid",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to POST /api/v0/certificate_authorities/some-certificate-authority-guid/activate"))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/certificate_authorities/some-certificate-authority-guid/activate"),
						ghttp.VerifyContentType("application/json"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
					GUID: "some-certificate-authority-guid",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request failed: unexpected response"))
			})
		})
	})

	Describe("DeleteCertificateAuthority", func() {
		It("deletes a certificate authority", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/certificate_authorities/some-certificate-authority-guid"),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
				GUID: "some-certificate-authority-guid",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		When("the client cannot make a request", func() {
			It("returns an error", func() {
				client.Close()

				err := service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
					GUID: "some-certificate-authority-guid",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not send api request to DELETE /api/v0/certificate_authorities/some-certificate-authority-guid"))
			})
		})

		When("Ops Manager returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/certificate_authorities/some-certificate-authority-guid"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err := service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
					GUID: "some-certificate-authority-guid",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request failed: unexpected response"))
			})
		})
	})
})
