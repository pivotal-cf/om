package api_test

import (
	"fmt"
	"net/http"
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
		It("returns expiring licenses for staged products", func() {
			expiryDate := formatDate(daysFromNow(20))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-fa24570b6a6e8940ab57",
							"guid": "cf-fa24570b6a6e8940ab57",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Ops Manager: Example Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-fa24570b6a6e8940ab57"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("does not return staged licenses that expire after the specified window", func() {
			expiryDate := formatDate(daysFromNow(60))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-fa24570b6a6e8940ab57",
							"guid": "cf-fa24570b6a6e8940ab57",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Ops Manager: Example Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(0))
		})

		It("returns expiring licenses for deployed products", func() {
			expiryDate := formatDate(daysFromNow(20))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-fa24570b6a6e8940ab57",
							"guid": "cf-fa24570b6a6e8940ab57",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Ops Manager: Example Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							],
							"stale": {
								"parent_products_deployed_more_recently": []
							}
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", false, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-fa24570b6a6e8940ab57"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("does not return deployed licenses that expire after the specified window", func() {
			expiryDate := formatDate(daysFromNow(60))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-fa24570b6a6e8940ab57",
							"guid": "cf-fa24570b6a6e8940ab57",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Ops Manager: Example Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							],
							"stale": {
								"parent_products_deployed_more_recently": []
							}
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", false, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(0))
		})

		It("returns expiring licenses from both staged and deployed products when neither flag is specified", func() {
			stagedExpiryDate := formatDate(daysFromNow(15))
			deployedExpiryDate := formatDate(daysFromNow(25))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-staged-guid",
							"guid": "cf-staged-guid",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Staged Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Staged product",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, stagedExpiryDate)),
				),
			)

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-deployed-guid",
							"guid": "cf-deployed-guid",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Deployed Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Deployed product",
									"product_version": "1.2.3.4"
								}
							],
							"stale": {
								"parent_products_deployed_more_recently": []
							}
						}
					]`, deployedExpiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", false, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(2))

			stagedLicense := findLicenseByGUID(licenses, "cf-staged-guid")
			Expect(stagedLicense).NotTo(BeNil())
			Expect(stagedLicense.ProductName).To(Equal("cf"))
			expectedStagedTime, _ := time.Parse("2006-01-02", stagedExpiryDate)
			Expect(stagedLicense.ExpiresAt).To(Equal(expectedStagedTime))

			deployedLicense := findLicenseByGUID(licenses, "cf-deployed-guid")
			Expect(deployedLicense).NotTo(BeNil())
			Expect(deployedLicense.ProductName).To(Equal("cf"))
			expectedDeployedTime, _ := time.Parse("2006-01-02", deployedExpiryDate)
			Expect(deployedLicense.ExpiresAt).To(Equal(expectedDeployedTime))
		})

		It("returns expiring licenses from both staged and deployed products when both flags are specified", func() {
			stagedExpiryDate := formatDate(daysFromNow(15))
			deployedExpiryDate := formatDate(daysFromNow(25))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-staged-guid",
							"guid": "cf-staged-guid",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Staged Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Staged product",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, stagedExpiryDate)),
				),
			)

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-deployed-guid",
							"guid": "cf-deployed-guid",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Deployed Licensed Product",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Deployed product",
									"product_version": "1.2.3.4"
								}
							],
							"stale": {
								"parent_products_deployed_more_recently": []
							}
						}
					]`, deployedExpiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(2))

			stagedLicense := findLicenseByGUID(licenses, "cf-staged-guid")
			Expect(stagedLicense).NotTo(BeNil())
			Expect(stagedLicense.ProductName).To(Equal("cf"))
			expectedStagedTime, _ := time.Parse("2006-01-02", stagedExpiryDate)
			Expect(stagedLicense.ExpiresAt).To(Equal(expectedStagedTime))

			deployedLicense := findLicenseByGUID(licenses, "cf-deployed-guid")
			Expect(deployedLicense).NotTo(BeNil())
			Expect(deployedLicense.ProductName).To(Equal("cf"))
			expectedDeployedTime, _ := time.Parse("2006-01-02", deployedExpiryDate)
			Expect(deployedLicense.ExpiresAt).To(Equal(expectedDeployedTime))
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

		It("correctly filters licenses using weeks as the time unit", func() {
			expiryDate := formatDate(daysFromNow(21))
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-weeks-test",
							"guid": "cf-weeks-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 3 weeks",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("2w", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(0))

			client.Reset()

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-weeks-test",
							"guid": "cf-weeks-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 3 weeks",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err = service.ListExpiringLicenses("4w", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].GUID).To(Equal("cf-weeks-test"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("correctly filters licenses using months as the time unit", func() {
			expiryDate := formatDate(time.Now().AddDate(0, 2, 0))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-months-test",
							"guid": "cf-months-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 2 months",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("1m", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(0))

			client.Reset()

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-months-test",
							"guid": "cf-months-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 2 months",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err = service.ListExpiringLicenses("3m", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].GUID).To(Equal("cf-months-test"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("correctly filters licenses using years as the time unit", func() {
			expiryDate := formatDate(time.Now().AddDate(0, 9, 0))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-years-test",
							"guid": "cf-years-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 9 months",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("6m", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(0))

			client.Reset()

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-years-test",
							"guid": "cf-years-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring in 9 months",
							"service_broker": false,
							"bosh_read_creds": false,
							"license_metadata": [
								{
									"property_reference": ".properties.license_key",
									"expiry": "%s",
									"product_name": "Some product!",
									"product_version": "1.2.3.4"
								}
							]
						}
					]`, expiryDate)),
				),
			)

			licenses, err = service.ListExpiringLicenses("1y", true, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].GUID).To(Equal("cf-years-test"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

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

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-multiple-licenses"))

			expectedTime, _ := time.Parse("2006-01-02", expiringDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
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

		It("deduplicates products that appear in both staged and deployed states", func() {
			expiryDate := formatDate(daysFromNow(20))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-duplicate-test",
							"guid": "cf-duplicate-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product in both states",
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
					]`, expiryDate)),
				),
			)

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-duplicate-test",
							"guid": "cf-duplicate-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product in both states",
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
					]`, expiryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, true)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-duplicate-test"))

			expectedTime, _ := time.Parse("2006-01-02", expiryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("correctly handles licenses expiring exactly on the boundary date", func() {
			boundaryDate := formatDate(daysFromNow(30))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-boundary-test",
							"guid": "cf-boundary-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring exactly on boundary",
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
					]`, boundaryDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-boundary-test"))

			expectedTime, _ := time.Parse("2006-01-02", boundaryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))

			client.Reset()
			boundaryDate = formatDate(daysFromNow(14))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-boundary-test",
							"guid": "cf-boundary-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product expiring exactly on boundary",
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
					]`, boundaryDate)),
				),
			)

			licenses, err = service.ListExpiringLicenses("2w", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-boundary-test"))

			expectedTime, _ = time.Parse("2006-01-02", boundaryDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})

		It("correctly handles licenses that have already expired", func() {
			expiredDate := formatDate(daysFromNow(-5))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-expired-test",
							"guid": "cf-expired-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product with expired license",
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
					]`, expiredDate)),
				),
			)

			licenses, err := service.ListExpiringLicenses("30d", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-expired-test"))

			expectedTime, _ := time.Parse("2006-01-02", expiredDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))

			client.Reset()
			expiredDate = formatDate(time.Now().AddDate(-1, 0, 0))

			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`[
						{
							"installation_name": "cf-long-expired-test",
							"guid": "cf-long-expired-test",
							"type": "cf",
							"product_version": "1.0-build.0",
							"label": "Product with long expired license",
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
					]`, expiredDate)),
				),
			)

			licenses, err = service.ListExpiringLicenses("30d", true, false)
			Expect(err).NotTo(HaveOccurred())

			Expect(licenses).To(HaveLen(1))
			Expect(licenses[0].ProductName).To(Equal("cf"))
			Expect(licenses[0].GUID).To(Equal("cf-long-expired-test"))

			expectedTime, _ = time.Parse("2006-01-02", expiredDate)
			Expect(licenses[0].ExpiresAt).To(Equal(expectedTime))
		})
	})
})

func daysFromNow(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func findLicenseByGUID(licenses []api.ExpiringLicenseOutPut, guid string) *api.ExpiringLicenseOutPut {
	for _, license := range licenses {
		if license.GUID == guid {
			return &license
		}
	}
	return nil
}
