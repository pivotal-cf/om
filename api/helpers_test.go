package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("Available Products", func() {
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

	Describe("CheckProductAvailability", func() {
		When("the product is available", func() {
			It("is true", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"name": "available-product",
							"product_version": "available-version"
						}]`),
					),
				)

				available, err := service.CheckProductAvailability("available-product", "available-version")
				Expect(err).ToNot(HaveOccurred())

				Expect(available).To(BeTrue())
			})
		})

		When("the product is unavailable", func() {
			It("is false", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/available_products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"name": "available-product",
							"product_version": "available-version"
						}]`),
					),
				)

				available, err := service.CheckProductAvailability("unavailable-product", "available-version")
				Expect(err).ToNot(HaveOccurred())

				Expect(available).To(BeFalse())
			})
		})

		When("the client can't connect to the server", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.CheckProductAvailability("", "")
				Expect(err).To(MatchError(ContainSubstring("could not make api request")))
			})
		})
	})

	Describe("RunningInstallation", func() {
		It("returns only the running installation on the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installations"),
					ghttp.RespondWith(http.StatusOK, `{
						"installations": [{
							"finished_at": null,
							"status": "running",
							"id": 3
						}, {
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"status": "succeeded",
							"id": 2
						}]
					}`),
				),
			)

			output, err := service.RunningInstallation()

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{
				ID:         3,
				Status:     "running",
				FinishedAt: parseTime(nil),
			}))
		})

		When("there are no installations", func() {
			It("returns a zero value installation", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusOK, `{"installations": []}`),
					),
				)

				output, err := service.RunningInstallation()

				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))
			})
		})

		When("there is no running installation", func() {
			It("returns a zero value installation", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusOK, `{
							"installations": [{
									"user_name": "admin",
									"finished_at": "2017-05-25T00:10:00.303Z",
									"status": "succeeded",
									"id": 3
								}, {
									"user_name": "admin",
									"finished_at": "2017-05-24T23:55:56.106Z",
									"status": "succeeded",
									"id": 2
							}]
						}`),
					),
				)

				output, err := service.RunningInstallation()

				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))
			})
		})

		When("only an earlier installation is listed in the running state", func() {
			It("does not consider the earlier installation to be running", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusOK, `{
							"installations": [{
								"finished_at": null,
								"status": "succeeded",
								"id": 3
							}, {
								"user_name": "admin",
								"finished_at": "2017-07-05T00:39:32.123Z",
								"status": "running",
								"id": 2
							}]
						}`),
					),
				)

				output, err := service.RunningInstallation()

				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))
			})
		})

		When("the client has an error during the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.RunningInstallation()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installations endpoint: could not send api request to GET /api/v0/installations")))
			})
		})

		When("the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.RunningInstallation()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the json cannot be decoded", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installations"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.RunningInstallation()
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})

	Describe("GetStagedProductByName", func() {
		It("Find product by product name", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[
						{"installation_name":"p-bosh","guid":"some-product-id","type":"some-product-name","product_version":"1.10.0.0"},
						{"installation_name":"cf-15b22d1810a034ea3aca","guid":"cf-15b22d1810a034ea3aca","type":"cf","product_version":"1.10.0-build.177"},
						{"installation_name":"p-isolation-segment-0ab7a3616c32a441a115","guid":"p-isolation-segment-0ab7a3616c32a441a115","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
					]`),
				),
			)

			finderOutput, err := service.GetStagedProductByName("some-product-name")
			Expect(err).ToNot(HaveOccurred())
			Expect(finderOutput.Product.GUID).To(Equal("some-product-id"))
		})

		Context("Failed to list staged products", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetStagedProductByName("some-product-name")
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged products response")))
			})
		})

		Context("Target product not in staged product list", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[
								{"installation_name":"cf-15b22d1810a034ea3aca","guid":"cf-15b22d1810a034ea3aca","type":"cf","product_version":"1.10.0-build.177"},
								{"installation_name":"p-isolation-segment-0ab7a3616c32a441a115","guid":"p-isolation-segment-0ab7a3616c32a441a115","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
							]`),
					),
				)

				_, err := service.GetStagedProductByName("some-product-name")
				Expect(err).To(MatchError(ContainSubstring("could not find product \"some-product-name\"")))
			})
		})
	})
})
