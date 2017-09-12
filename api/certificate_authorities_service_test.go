package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("CertificateAuthoritiesService", func() {
	var (
		client  *fakes.HttpClient
		service api.CertificateAuthoritiesService
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.NewCertificateAuthoritiesService(client)
	})

	Describe("List", func() {
		It("returns a slice of CAs", func() {
			var path string
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				path = req.URL.Path

				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
"certificate_authorities": [
	{
		"guid": "some-guid",
		"issuer": "some-issuer",
		"created_on": "2017-01-09",
		"expires_on": "2021-01-09",
		"active": true,
		"cert_pem": "some-cert-pem"
	},
	{
		"guid": "some-guid-2",
		"issuer": "another-issuer",
		"created_on": "2017-09-09",
		"expires_on": "2021-10-02",
		"active": false,
		"cert_pem": "some-other-cert-pem"
	}
]
}`)),
				}, nil
			}

			output, err := service.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(output.CAs).To(ConsistOf([]api.CA{
				{
					GUID:      "some-guid",
					Issuer:    "some-issuer",
					CreatedOn: "2017-01-09",
					ExpiresOn: "2021-01-09",
					Active:    true,
					CertPEM:   "some-cert-pem",
				},
				{
					GUID:      "some-guid-2",
					Issuer:    "another-issuer",
					CreatedOn: "2017-09-09",
					ExpiresOn: "2021-10-02",
					Active:    false,
					CertPEM:   "some-other-cert-pem",
				},
			}))

			Expect(path).To(Equal("/api/v0/certificate_authorities"))
		})

		Context("failure cases", func() {
			Context("when the client cannot make a request", func() {
				It("returns an error", func() {
					client.DoReturns(nil, errors.New("client do errored"))

					_, err := service.List()
					Expect(err).To(MatchError("client do errored"))
				})
			})

			Context("when the response body cannot be parsed", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(strings.NewReader(`%%%%`)),
						}, nil
					}

					_, err := service.List()
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
