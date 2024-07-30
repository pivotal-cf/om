package api_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("Expiring Certificates", func() {
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

	When("getting a list of expiring certificates", func() {
		It("supports a expiration range and returns a detailed response", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/certificates", "expires_within=3d"),
					ghttp.RespondWith(http.StatusOK, `{
				  		"certificates": [{
					 		"issuer": "/CN=opsmgr-bosh-dns-tls-ca",
					 		"valid_from": "2018-08-10T21:07:37Z",
					 		"valid_until": "2022-08-09T21:07:37Z",
					 		"configurable": false,
					 		"property_reference": null,
					 		"property_type": null,
					 		"product_guid": null,
					 		"location": "credhub",
					 		"variable_path": "/opsmgr/bosh_dns/tls_ca"
						}]
					}`),
				),
			)

			expiresWithin := "3d"
			certs, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).ToNot(HaveOccurred())

			fromTime, err := time.Parse(time.RFC3339, "2018-08-10T21:07:37Z")
			Expect(err).ToNot(HaveOccurred())
			toTime, err := time.Parse(time.RFC3339, "2022-08-09T21:07:37Z")
			Expect(err).ToNot(HaveOccurred())

			Expect(certs).To(Equal([]api.ExpiringCertificate{
				{
					Issuer:            "/CN=opsmgr-bosh-dns-tls-ca",
					ValidFrom:         fromTime,
					ValidUntil:        toTime,
					Configurable:      false,
					PropertyReference: "",
					PropertyType:      "",
					ProductGUID:       "",
					Location:          "credhub",
					VariablePath:      "/opsmgr/bosh_dns/tls_ca",
				},
			}))
		})
	})

	DescribeTable("time durations are passed", func(expiresWithin string, expectedTime string) {
		client.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v0/deployed/certificates", fmt.Sprintf("expires_within=%s", expectedTime)),
				ghttp.RespondWith(http.StatusOK, `{"certificates": []}`),
			),
		)

		_, err := service.ListExpiringCertificates(expiresWithin)
		Expect(err).ToNot(HaveOccurred())
	},
		Entry("days", "2d", "2d"),
		Entry("weeks", "1w", "1w"),
		Entry("months", "1m", "1m"),
		Entry("years", "1y", "1y"),
	)

	When("the api returns an error", func() {
		It("returns the error", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/certificates", "expires_within=3d"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			expiresWithin := "3d"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the HTTP client returns an error", func() {
		It("returns the error", func() {
			client.Close()

			expiresWithin := "3d"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the response can't be unmarshaled", func() {
		It("returns the error", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/certificates", "expires_within=3d"),
					ghttp.RespondWith(http.StatusOK, `invalid-json`),
				),
			)

			expiresWithin := "3d"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})
})
