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
			Expect(err.Error()).To(ContainSubstring("could not get staged products"))
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
