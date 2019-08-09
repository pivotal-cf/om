package api_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Expiring Certificates", func() {
	var (
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			Client: client,
		})
	})

	When("getting a list of expiring certificates", func() {
		It("supports a expiration range and returns a detailed response", func() {
			client.DoStub = func(request *http.Request) (response *http.Response, e error) {
				Expect(request.URL.Path).To(Equal("/api/v0/deployed/certificates"))
				Expect(request.URL.RawQuery).To(Equal("expires_within=3h"))

				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`
						{
						  "certificates": [
							{
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
						}
					`)),
				}, nil
			}

			expiresWithin := "3h"
			certs, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).NotTo(HaveOccurred())

			fromTime, err := time.Parse(time.RFC3339, "2018-08-10T21:07:37Z")
			Expect(err).NotTo(HaveOccurred())
			toTime, err := time.Parse(time.RFC3339, "2022-08-09T21:07:37Z")
			Expect(err).NotTo(HaveOccurred())

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
		client.DoStub = func(request *http.Request) (response *http.Response, e error) {
			Expect(request.URL.Path).To(Equal("/api/v0/deployed/certificates"))
			Expect(request.URL.RawQuery).To(Equal(fmt.Sprintf("expires_within=%s", expectedTime)))

			return &http.Response{StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`
					{
					  "certificates": []
					}
				`)),
			}, nil
		}

		_, err := service.ListExpiringCertificates(expiresWithin)
		Expect(err).NotTo(HaveOccurred())
	},
		Entry("days", "2d", "2d"),
		Entry("weeks", "1w", "1w"),
		Entry("months", "1m", "1m"),
		Entry("years", "1y", "1y"),
	)

	When("the api returns an error", func() {
		It("returns the error", func() {
			client.DoStub = func(request *http.Request) (response *http.Response, e error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, nil
			}

			expiresWithin := "3h"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the HTTP client returns an error", func() {
		It("returns the error", func() {
			client.DoStub = func(request *http.Request) (response *http.Response, e error) {
				return nil, fmt.Errorf("some error")
			}

			expiresWithin := "3h"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})

	When("the response can't be unmarshaled", func() {
		It("returns the error", func() {
			client.DoStub = func(request *http.Request) (response *http.Response, e error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`
						{"invalid-json'
					`))}, nil
			}

			expiresWithin := "3h"
			_, err := service.ListExpiringCertificates(expiresWithin)
			Expect(err).To(HaveOccurred())
		})
	})
})
