package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
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

			valMap, ok := output.Credential.Value.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(valMap["private_key_pem"]).To(Equal("some-private-key"))
			Expect(valMap["cert_pem"]).To(Equal("some-cert-pem"))
		})

		It("fetches a generated_credentials reference with nested output value", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials/.properties.generated-cred"),
					ghttp.RespondWith(http.StatusOK, `{
						"credential":{
							"type": "generated_credentials",
							"credential": true,
							"value":{
								"output": {
									"host": "10.0.0.1",
									"port": 8080,
									"username": "admin"
								}
							}
						}
					}`),
				),
			)

			output, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
				DeployedGUID:        "some-deployed-product-guid",
				CredentialReference: ".properties.generated-cred",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Credential.Type).To(Equal("generated_credentials"))

			valMap, ok := output.Credential.Value.(map[string]interface{})
			Expect(ok).To(BeTrue())

			outputVal, ok := valMap["output"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(outputVal["host"]).To(Equal("10.0.0.1"))
			Expect(outputVal["port"]).To(Equal(float64(8080)))
			Expect(outputVal["username"]).To(Equal("admin"))
		})

		It("fetches a generated_credentials reference with simple string output", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials/.properties.generated-string-cred"),
					ghttp.RespondWith(http.StatusOK, `{
						"credential":{
							"type": "generated_credentials",
							"credential": true,
							"value":{
								"output": "some-generated-string-value"
							}
						}
					}`),
				),
			)

			output, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
				DeployedGUID:        "some-deployed-product-guid",
				CredentialReference: ".properties.generated-string-cred",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Credential.Type).To(Equal("generated_credentials"))

			valMap, ok := output.Credential.Value.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(valMap["output"]).To(Equal("some-generated-string-value"))
		})

		It("fetches a generated_credentials reference with empty output", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products/some-deployed-product-guid/credentials/.properties.generated-empty-cred"),
					ghttp.RespondWith(http.StatusOK, `{
						"credential":{
							"type": "generated_credentials",
							"credential": true,
							"value":{
								"output": ""
							}
						}
					}`),
				),
			)

			output, err := service.GetDeployedProductCredential(api.GetDeployedProductCredentialInput{
				DeployedGUID:        "some-deployed-product-guid",
				CredentialReference: ".properties.generated-empty-cred",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Credential.Type).To(Equal("generated_credentials"))

			valMap, ok := output.Credential.Value.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(valMap["output"]).To(Equal(""))
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
