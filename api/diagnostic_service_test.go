package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Diagnostic Report", func() {
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

	Describe("Ops Man pre 2.6", func() {
		Describe("DiagnosticReport", func() {
			It("returns a diagnostic report", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
						ghttp.RespondWith(http.StatusOK, pre2_6Report),
					),
				)

				report, err := service.GetDiagnosticReport()
				Expect(err).ToNot(HaveOccurred())

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

			When("an error occurs", func() {
				When("the server returns a 500", func() {
					It("returns a DiagnosticReportUnavailable error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusInternalServerError, `{}`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).To(BeAssignableToTypeOf(api.DiagnosticReportUnavailable{}))
					})
				})

				When("the client fails before the request", func() {
					It("returns an error", func() {
						client.Close()

						_, err := service.GetDiagnosticReport()
						Expect(err).To(MatchError(ContainSubstring("could not make api request to diagnostic_report endpoint: could not send api request to GET /api/v0/diagnostic_report")))
					})
				})

				When("the server returns a non-2XX status", func() {
					It("returns an error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusTeapot, `{}`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).ToNot(MatchError("request failed: unexpected response"))
					})
				})

				When("invalid json is returned", func() {
					It("returns an error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusOK, `%%%`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).ToNot(MatchError("invalid json received from server"))
					})
				})
			})
		})
	})

	Describe("Ops Man post 2.6", func() {
		Describe("DiagnosticReport", func() {
			It("returns a diagnostic report", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
						ghttp.RespondWith(http.StatusOK, post2_6Report),
					),
				)

				report, err := service.GetDiagnosticReport()
				Expect(err).ToNot(HaveOccurred())

				Expect(report.InfrastructureType).To(Equal("azure"))
				Expect(report.AvailableStemcells).To(Equal([]api.Stemcell{
					{
						Filename: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
						OS:       "ubuntu-trusty",
						Version:  "3263.8",
					},
				}))
				Expect(report.StagedProducts).To(Equal([]api.DiagnosticProduct{
					{
						Name:    "p-bosh",
						Version: "1.8.8.0",
						Stemcells: []api.Stemcell{
							{
								Filename: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
								OS:       "ubuntu-trusty",
								Version:  "3263.8",
							},
						},
					},
					{
						Name:    "gcp-service-broker",
						Version: "2.0.1",
						Stemcells: []api.Stemcell{
							{
								Filename: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
								OS:       "ubuntu-trusty",
								Version:  "3263.8",
							},
						},
					},
					{
						Name:    "gitlab-ee",
						Version: "1.0.1",
					},
				}))

				Expect(report.DeployedProducts).To(Equal([]api.DiagnosticProduct{
					{
						Name:    "p-bosh",
						Version: "1.8.8.0",
						Stemcells: []api.Stemcell{
							{
								Filename: "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
								OS:       "ubuntu-trusty",
								Version:  "3263.8",
							},
						},
					},
				}))
			})

			When("an error occurs", func() {
				When("the server returns a 500", func() {
					It("returns a DiagnosticReportUnavailable error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusInternalServerError, `{}`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).To(BeAssignableToTypeOf(api.DiagnosticReportUnavailable{}))
					})
				})

				When("the client fails before the request", func() {
					It("returns an error", func() {
						client.Close()

						_, err := service.GetDiagnosticReport()
						Expect(err).To(MatchError(ContainSubstring("could not make api request to diagnostic_report endpoint: could not send api request to GET /api/v0/diagnostic_report")))
					})
				})

				When("the server returns a non-2XX status", func() {
					It("returns an error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusTeapot, `{}`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).ToNot(MatchError("request failed: unexpected response"))
					})
				})

				When("invalid json is returned", func() {
					It("returns an error", func() {
						client.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
								ghttp.RespondWith(http.StatusOK, `%%%`),
							),
						)

						_, err := service.GetDiagnosticReport()
						Expect(err).ToNot(MatchError("invalid json received from server"))
					})
				})
			})
		})
	})

	Describe("Full Diagnostic Report", func() {
		It("returns the full report", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/diagnostic_report"),
					ghttp.RespondWith(http.StatusOK, post2_6Report),
				),
			)

			report, err := service.GetDiagnosticReport()
			Expect(err).ToNot(HaveOccurred())

			Expect(report.FullReport).To(Equal(post2_6Report))
		})
	})
})

const pre2_6Report = `{
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
const post2_6Report = `{
  "infrastructure_type": "azure",
  "available_stemcells": [
    {
      "filename": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
      "os": "ubuntu-trusty",
      "version": "3263.8"
    }
  ],
  "added_products": {
    "deployed": [
      {
        "name": "p-bosh",
        "version": "1.8.8.0",
        "stemcells": [
          {
            "filename": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
            "os": "ubuntu-trusty",
            "version": "3263.8"
          }
        ]
      }
    ],
    "staged": [
      {
        "name": "p-bosh",
        "version": "1.8.8.0",
        "stemcells": [
          {
            "filename": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
            "os": "ubuntu-trusty",
            "version": "3263.8"
          }
        ]
      },
      {
        "name": "gcp-service-broker",
        "version": "2.0.1",
        "stemcells": [
          {
            "filename": "light-bosh-stemcell-3263.8-aws-xen-hvm-ubuntu-trusty-go_agent.tgz",
            "os": "ubuntu-trusty",
            "version": "3263.8"
          }
        ]
      },
      {
        "name": "gitlab-ee",
        "version": "1.0.1"
      }
    ]
  }
}`
