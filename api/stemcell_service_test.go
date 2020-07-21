package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("StemcellService", func() {
	var (
		server  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		client := httpClient{
			server.URL(),
		}

		service = api.New(api.ApiInput{
			Client:         client,
			UnauthedClient: client,
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ListStemcells", func() {
		It("makes a request to list the stemcells", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
					ghttp.RespondWith(http.StatusOK, `{
						"products": [{
							"guid": "some-guid",
							"staged_stemcell_version": "1234.5",
							"identifier": "some-product",
							"available_stemcell_versions": [
								"1234.5", "1234.6"
							]
						}]
					}`),
				),
			)

			output, err := service.ListStemcells()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "some-guid",
						StagedForDeletion:     false,
						StagedStemcellVersion: "1234.5",
						ProductName:           "some-product",
						AvailableVersions: []string{
							"1234.5",
							"1234.6",
						},
					},
				},
			}))
		})

		When("invalid JSON is returned", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of value")))
			})
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_assignments")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("AssignStemcells", func() {
		It("makes a request to assign the stemcells", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_assignments"),
					ghttp.VerifyJSON(`{
						"products": [{
							"guid": "some-guid",
							"staged_stemcell_version": "1234.6"
						}]
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.AssignStemcell(api.ProductStemcells{
				Products: []api.ProductStemcell{{
					GUID:                  "some-guid",
					StagedStemcellVersion: "1234.6",
				}},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("the server errors before the request", func() {
			It("returns an error", func() {
				server.Close()

				err := service.AssignStemcell(api.ProductStemcells{
					Products: []api.ProductStemcell{{
						GUID:                  "some-guid",
						StagedStemcellVersion: "1234.6",
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("could not send api request to PATCH /api/v0/stemcell_assignments")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_assignments"),
						ghttp.VerifyJSON(`{
							"products": [{
								"guid": "some-guid",
								"staged_stemcell_version": "1234.6"
							}]
						}`),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.AssignStemcell(api.ProductStemcells{
					Products: []api.ProductStemcell{{
						GUID:                  "some-guid",
						StagedStemcellVersion: "1234.6",
					}},
				})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("CheckStemcellAvailability", func() {
		When("the diagnostic report exists with stemcells", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, `
				{
					"stemcells": ["light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz"],
					"available_stemcells": [{
						"filename": "light-bosh-stemcell-621.80-google-kvm-ubuntu-xenial-go_agent.tgz"
					}]
				}
				`),
				)
			})

			When("the version of the OpsManager cannot be determined", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{}`),
					)

					_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not determine version"))
				})
			})

			When("the version cannot be gotten", func() {
				It("returns an error", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusNotFound, nil),
					)

					_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("cannot retrieve version"))
				})
			})

			When("the stemcell does not exists", func() {
				It("returns false", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.4.3"}}`),
					)

					found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).NotTo(HaveOccurred())

					Expect(found).To(BeFalse())
				})

				When("the OpsManager is 2.6+", func() {
					It("returns false", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.77-google-kvm-ubuntu-xenial-go_agent.tgz")
						Expect(err).NotTo(HaveOccurred())

						Expect(found).To(BeFalse())
					})
				})
			})

			When("the stemcell already exists", func() {
				It("returns true", func() {
					server.AppendHandlers(
						ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.4.3"}}`),
					)

					found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz")
					Expect(err).NotTo(HaveOccurred())

					Expect(found).To(BeTrue())
				})

				When("the OpsMan 2.6+", func() {
					It("returns true", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusOK, `{"info":{"version":"2.6.3"}}`),
						)

						found, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.80-google-kvm-ubuntu-xenial-go_agent.tgz")
						Expect(err).NotTo(HaveOccurred())

						Expect(found).To(BeTrue())
					})
				})
			})
		})

		When("the diagnostic report fails", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusNotFound, nil),
				)
				_, err := service.CheckStemcellAvailability("light-bosh-stemcell-621.79-google-kvm-ubuntu-xenial-go_agent.tgz")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
