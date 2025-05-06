package api_test

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("ExpiringLicenseService", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("ListExpiringLicenses", func() {
		Describe("staged products", func() {
			It("returns expiring licenses for staged products", func() {
				expiryDate := formatDate(daysFromNow(20))
				guid := "cf-fa24570b6a6e8940ab57"
				label := "Ops Manager: Example Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiryDate, "staged")
			})

			It("does not return staged licenses that expire after the specified window", func() {
				expiryDate := formatDate(daysFromNow(60))
				guid := "cf-fa24570b6a6e8940ab57"
				label := "Ops Manager: Example Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())

				Expect(licenses).To(HaveLen(0))
			})

			It("does not return staged Products that have perpetual license", func() {
				expiryDate := ""
				guid := "cf-fa24570b6a6e8940ab57"
				label := "Ops Manager: Example Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/cf-fa24570b6a6e8940ab57/properties"),
						ghttp.RespondWith(http.StatusOK, createStagedProductPropertiesResponse("test-license-key")),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())

				Expect(licenses).To(HaveLen(0))
			})

			It("returns an error when the API call to get staged products fails", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"error": "server error"}`),
					),
				)

				_, err := service.ListExpiringLicenses("30d", true, false)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not get staged products"))
				Expect(err.Error()).To(ContainSubstring("could not make a call to ListStagedProducts api"))
			})
		})

		Describe("deployed products", func() {
			It("returns expiring licenses for deployed products", func() {
				expiryDate := formatDate(daysFromNow(20))
				guid := "cf-fa24570b6a6e8940ab57"
				label := "Ops Manager: Example Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", false, true)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiryDate, "deployed")
			})

			It("does not return deployed licenses that expire after the specified window", func() {
				expiryDate := formatDate(daysFromNow(60))
				guid := "cf-fa24570b6a6e8940ab57"
				label := "Ops Manager: Example Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", false, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(licenses).To(HaveLen(0))
			})

			It("returns an error when the API call to get deployed products fails", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"error": "server error"}`),
					),
				)

				_, err := service.ListExpiringLicenses("30d", false, true)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not get deployed products"))
				Expect(err.Error()).To(ContainSubstring("could not make a call to ListDeployedProducts api"))
			})
		})

		Describe("combined staged and deployed products", func() {
			It("returns expiring licenses from both staged and deployed products when neither flag is specified", func() {
				stagedExpiryDate := formatDate(daysFromNow(15))
				deployedExpiryDate := formatDate(daysFromNow(25))
				stagedGUID := "cf-staged-guid"
				deployedGUID := "cf-deployed-guid"
				stagedLabel := "Staged Licensed Product"
				deployedLabel := "Deployed Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(deployedGUID, deployedExpiryDate, deployedLabel)),
					),
				)
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(stagedGUID, stagedExpiryDate, stagedLabel)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", false, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(2))
				expectLicenseDetails(findLicenseByGUID(licenses, stagedGUID), stagedGUID, stagedExpiryDate, "staged")
				expectLicenseDetails(findLicenseByGUID(licenses, deployedGUID), deployedGUID, deployedExpiryDate, "deployed")
			})

			It("returns expiring licenses from both staged and deployed products when both flags are specified", func() {
				stagedExpiryDate := formatDate(daysFromNow(15))
				deployedExpiryDate := formatDate(daysFromNow(25))
				stagedGUID := "cf-staged-guid"
				deployedGUID := "cf-deployed-guid"
				stagedLabel := "Staged Licensed Product"
				deployedLabel := "Deployed Licensed Product"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(deployedGUID, deployedExpiryDate, deployedLabel)),
					),
				)
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(stagedGUID, stagedExpiryDate, stagedLabel)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(2))

				// Find licenses by their state
				var stagedLicense, deployedLicense *api.ExpiringLicenseOutput
				for i := range licenses {
					if slices.Equal(licenses[i].ProductState, []string{"staged"}) {
						stagedLicense = &licenses[i]
					} else if slices.Equal(licenses[i].ProductState, []string{"deployed"}) {
						deployedLicense = &licenses[i]
					}
				}

				// Verify staged license
				Expect(stagedLicense).NotTo(BeNil())
				Expect(stagedLicense.GUID).To(Equal(stagedGUID))
				Expect(stagedLicense.ProductState).To(Equal([]string{"staged"}))
				expectedStagedTime, _ := time.Parse("2006-01-02", stagedExpiryDate)
				Expect(stagedLicense.ExpiresAt).To(Equal(expectedStagedTime))

				// Verify deployed license
				Expect(deployedLicense).NotTo(BeNil())
				Expect(deployedLicense.GUID).To(Equal(deployedGUID))
				Expect(deployedLicense.ProductState).To(Equal([]string{"deployed"}))
				expectedDeployedTime, _ := time.Parse("2006-01-02", deployedExpiryDate)
				Expect(deployedLicense.ExpiresAt).To(Equal(expectedDeployedTime))
			})
		})

		Describe("time window handling", func() {
			It("correctly filters licenses using weeks as the time unit", func() {
				expiryDate := formatDate(daysFromNow(21))
				guid := "cf-weeks-test"
				label := "Product expiring in 3 weeks"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("2w", true, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(0))

				client.Reset()

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err = service.ListExpiringLicenses("4w", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiryDate, "staged")
			})

			It("correctly filters licenses using months as the time unit", func() {
				expiryDate := formatDate(time.Now().AddDate(0, 2, 0))
				guid := "cf-months-test"
				label := "Product expiring in 2 months"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("1m", true, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(0))

				client.Reset()

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err = service.ListExpiringLicenses("3m", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiryDate, "staged")
			})

			It("correctly filters licenses using years as the time unit", func() {
				expiryDate := formatDate(time.Now().AddDate(0, 9, 0))
				guid := "cf-years-test"
				label := "Product expiring in 9 months"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("6m", true, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(0))

				client.Reset()

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
					),
				)

				licenses, err = service.ListExpiringLicenses("1y", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiryDate, "staged")
			})

			It("correctly handles licenses expiring exactly on the boundary date", func() {
				boundaryDate := formatDate(daysFromNow(30))
				guid := "cf-boundary-test"
				label := "Product expiring exactly on boundary"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, boundaryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, boundaryDate, "staged")

				client.Reset()
				boundaryDate = formatDate(daysFromNow(14))

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, boundaryDate, label)),
					),
				)

				licenses, err = service.ListExpiringLicenses("2w", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, boundaryDate, "staged")
			})
		})

		Describe("edge cases", func() {
			It("correctly handles products with multiple licenses (some expiring, some not)", func() {
				expiringDate := formatDate(daysFromNow(20))
				futureDate := formatDate(daysFromNow(60))

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
							{
								"installation_name": "cf-multiple-licenses",
								"guid": "cf-multiple-licenses",
								"type": "cf",
								"product_version": "1.0-build.0",
								"label": "Product with multiple licenses",
								"service_broker": false,
								"bosh_read_creds": false,
								"license_metadata": [
									{
										"property_reference": ".properties.license_key_1",
										"expiry": "%s",
										"product_name": "Expiring License",
										"product_version": "1.2.3.4"
									},
									{
										"property_reference": ".properties.license_key_2",
										"expiry": "%s",
										"product_name": "Future License",
										"product_version": "1.2.3.4"
									}
								]
							}
						]`, expiringDate, futureDate)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, "cf-multiple-licenses", expiringDate, "staged")
			})

			It("correctly handles products with no license metadata", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[
							{
								"installation_name": "cf-no-license",
								"guid": "cf-no-license",
								"type": "cf",
								"product_version": "1.0-build.0",
								"label": "Product without license",
								"service_broker": false,
								"bosh_read_creds": false,
								"license_metadata": []
							}
						]`),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(0))
			})

			It("correctly handles licenses that have already expired", func() {
				expiredDate := formatDate(daysFromNow(-5))
				guid := "cf-expired-test"
				label := "Product with expired license"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiredDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiredDate, "staged")

				client.Reset()
				expiredDate = formatDate(time.Now().AddDate(-1, 0, 0))
				guid = "cf-long-expired-test"
				label = "Product with long expired license"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiredDate, label)),
					),
				)

				licenses, err = service.ListExpiringLicenses("30d", true, false)
				Expect(err).NotTo(HaveOccurred())
				expectSingleLicense(licenses, guid, expiredDate, "staged")
			})

			It("correctly handles different licenses for the same product in staged and deployed states", func() {
				guid := "cf-same-product"
				stagedExpiryDate := formatDate(daysFromNow(15))
				deployedExpiryDate := formatDate(daysFromNow(25))
				label := "Same Product Different Licenses"

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
						ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(guid, deployedExpiryDate, label)),
					),
				)

				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, stagedExpiryDate, label)),
					),
				)

				licenses, err := service.ListExpiringLicenses("30d", true, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(licenses).To(HaveLen(2))

				var stagedLicense, deployedLicense *api.ExpiringLicenseOutput
				for i := range licenses {
					if slices.Equal(licenses[i].ProductState, []string{"staged"}) {
						stagedLicense = &licenses[i]
					} else if slices.Equal(licenses[i].ProductState, []string{"deployed"}) {
						deployedLicense = &licenses[i]
					}
				}

				Expect(stagedLicense).NotTo(BeNil())
				Expect(stagedLicense.GUID).To(Equal(guid))
				Expect(stagedLicense.ProductState).To(Equal([]string{"staged"}))
				expectedStagedTime, _ := time.Parse("2006-01-02", stagedExpiryDate)
				Expect(stagedLicense.ExpiresAt).To(Equal(expectedStagedTime))

				Expect(deployedLicense).NotTo(BeNil())
				Expect(deployedLicense.GUID).To(Equal(guid))
				Expect(deployedLicense.ProductState).To(Equal([]string{"deployed"}))
				expectedDeployedTime, _ := time.Parse("2006-01-02", deployedExpiryDate)
				Expect(deployedLicense.ExpiresAt).To(Equal(expectedDeployedTime))
			})
		})

		It("combines the output when the same license is configured for the same product in staged and deployed states", func() {
			guid := "cf-same-product"
			expiryDate := formatDate(daysFromNow(15))
			label := "Same Product Same Licenses"

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, createDeployedProductResponse(guid, expiryDate, label)),
				),
			)

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, createStagedProductResponse(guid, expiryDate, label)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(1))
			combinedLicense := licenses[0]

			Expect(combinedLicense).NotTo(BeNil())
			Expect(combinedLicense.GUID).To(Equal(guid))
			Expect(combinedLicense.ProductState).To(Equal([]string{"deployed", "staged"}))
			expectedStagedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(combinedLicense.ExpiresAt).To(Equal(expectedStagedTime))
		})
	})
})

func createStagedProductResponse(guid, expiryDate, label string) string {
	return fmt.Sprintf(`[
		{
			"installation_name": "%s",
			"guid": "%s",
			"type": "cf",
			"product_version": "1.0-build.0",
			"label": "%s",
			"service_broker": false,
			"bosh_read_creds": false,
			"license_metadata": [
				{
					"property_reference": ".properties.license_key",
					"expiry": "%s",
					"product_name": "Test Product",
					"product_version": "1.2.3.4"
				}
			]
		}
	]`, guid, guid, label, expiryDate)
}

func createDeployedProductResponse(guid, expiryDate, label string) string {
	return fmt.Sprintf(`[
		{
			"installation_name": "%s",
			"guid": "%s",
			"type": "cf",
			"product_version": "1.0-build.0",
			"label": "%s",
			"service_broker": false,
			"bosh_read_creds": false,
			"license_metadata": [
				{
					"property_reference": ".properties.license_key",
					"expiry": "%s",
					"product_name": "Test Product",
					"product_version": "1.2.3.4"
				}
			],
			"stale": {
				"parent_products_deployed_more_recently": []
			}
		}
	]`, guid, guid, label, expiryDate)
}

func createStagedProductPropertiesResponse(value string) string {

	return fmt.Sprintf(`{
		  "properties": {
			".properties.license_key": {
			  "type": "tanzu_license_key",
			  "configurable": true,
			  "credential": false,
			  "value": "%s"
			}
		  }
		}`, value)
}

func daysFromNow(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func findLicenseByGUID(licenses []api.ExpiringLicenseOutput, guid string) *api.ExpiringLicenseOutput {
	for _, license := range licenses {
		if license.GUID == guid {
			return &license
		}
	}
	return nil
}

func expectSingleLicense(licenses []api.ExpiringLicenseOutput, guid, expiryDate string, productState string) {
	Expect(licenses).To(HaveLen(1))
	expectLicenseDetails(&licenses[0], guid, expiryDate, productState)
}

func expectLicenseDetails(license *api.ExpiringLicenseOutput, guid, expiryDate string, productState string) {
	Expect(license).NotTo(BeNil())
	Expect(license.ProductName).To(Equal("cf"))
	Expect(license.GUID).To(Equal(guid))
	Expect(license.ProductState).To(Equal([]string{productState}))

	expectedTime, _ := time.Parse("2006-01-02", expiryDate)
	Expect(license.ExpiresAt).To(Equal(expectedTime))
}
