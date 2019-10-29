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

	Describe("ListStemcells", func() {
		It("makes a request to list the stemcells", func() {
			client.AppendHandlers(
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
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_assignments"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of value")))
			})
		})

		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListStemcells()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_assignments")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
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
			client.AppendHandlers(
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

		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

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
				client.AppendHandlers(
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
})
