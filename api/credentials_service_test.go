package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Credentials", func() {
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

	Describe("ListDeployedProductCredentials", func() {
		It("lists credential references", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials"),
					ghttp.RespondWith(http.StatusOK, `{
						"credentials": [
							".properties.some-credentials", 
							".my-job.some-credentials"
						]
					}`),
				),
			)

			output, err := service.ListDeployedProductCredentials("some-deployed-product-guid")
			Expect(err).ToNot(HaveOccurred())

			Expect(output.Credentials).To(ConsistOf(
				[]string{
					".properties.some-credentials",
					".my-job.some-credentials",
				},
			))
		})

		Context("the client can't connect to the server", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListDeployedProductCredentials("invalid-product")
				Expect(err).To(MatchError(ContainSubstring("could not make api request")))
			})
		})

		When("the server won't fetch credential references", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-product-guid/credentials"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.ListDeployedProductCredentials("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("request failed")))
			})
		})

		When("the response is not JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListDeployedProductCredentials("some-deployed-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
			})
		})
	})

	Describe("GetDeployedProductCredential", func() {
		It("fetch a credential reference", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials/.properties.some-credentials"),
					ghttp.RespondWith(http.StatusOK, `{
						"credential":{
							"type": "rsa_cert_credentials",
							"credential": true,
							"value":{
								"private_key_pem": "some-private-key",
								"cert_pem": "some-cert-pem"
							}
						}
					}`),
				),
			)

			output, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
				DeployedGUID:        "some-deployed-product-guid",
				CredentialReference: ".properties.some-credentials",
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(output.Credential.Value["private_key_pem"]).To(Equal("some-private-key"))
			Expect(output.Credential.Value["cert_pem"]).To(Equal("some-cert-pem"))
		})

		When("the client can't connect to the server", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
					DeployedGUID: "invalid-product-guid",
				})
				Expect(err).To(MatchError(ContainSubstring("could not make api request")))
			})
		})

		When("the server won't fetch credential references", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/invalid-product-guid/credentials/"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
					DeployedGUID: "invalid-product-guid",
				})
				Expect(err).To(MatchError(ContainSubstring("request failed")))
			})
		})

		When("the response is not JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials/"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
					DeployedGUID: "some-deployed-product-guid",
				})
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
			})
		})
	})
})
