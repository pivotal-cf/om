package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const report = `{
  "infrastructure_type": "azure",
  "stemcells": ["light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"],
  "added_products": {
    "deployed": [
      {
        "name": "p-bosh",
        "version": "1.8.8.0",
        "stemcell": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
      }
    ],
    "staged": [
      {
        "name": "p-bosh",
        "version": "1.8.8.0",
        "stemcell": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
      },
      {
        "name": "gcp-service-broker",
        "version": "2.0.1",
        "stemcell": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"
      },
      {
        "name": "gitlab-ee",
        "version": "1.0.1"
      }
    ]
  }
}`

var _ = Describe("Diagnostic", func() {
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

	Describe("DiagnosticReport", func() {
		It("returns a diagnostic report", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(report)),
			}, nil)

			report, err := service.GetDiagnosticReport()
			Expect(err).NotTo(HaveOccurred())

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/diagnostic_report"))

			Expect(report.InfrastructureType).To(Equal("azure"))
			Expect(report.Stemcells).To(Equal([]string{"light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz"}))
			Expect(report.StagedProducts).To(Equal([]api.DiagnosticProduct{
				{
					Name:     "p-bosh",
					Version:  "1.8.8.0",
					Stemcell: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
				},
				{
					Name:     "gcp-service-broker",
					Version:  "2.0.1",
					Stemcell: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
				},
				{
					Name:    "gitlab-ee",
					Version: "1.0.1",
				},
			}))

			Expect(report.DeployedProducts).To(Equal([]api.DiagnosticProduct{
				{
					Name:     "p-bosh",
					Version:  "1.8.8.0",
					Stemcell: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the server returns a 500", func() {
				It("returns a DiagnosticReportUnavailable error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.GetDiagnosticReport()
					Expect(err).To(BeAssignableToTypeOf(api.DiagnosticReportUnavailable{}))
				})
			})

			Context("when the client fails before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))

					_, err := service.GetDiagnosticReport()
					Expect(err).To(MatchError("could not make api request to diagnostic_report endpoint: could not send api request to GET /api/v0/diagnostic_report: some error"))
				})
			})

			Context("when the server returns a non-2XX status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.GetDiagnosticReport()
					Expect(err).NotTo(MatchError("request failed: unexpected response"))
				})
			})

			Context("when invalid json is returned", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`$$$$$`)),
					}, nil)

					_, err := service.GetDiagnosticReport()
					Expect(err).NotTo(MatchError("invalid json received from server"))
				})
			})
		})
	})
})
