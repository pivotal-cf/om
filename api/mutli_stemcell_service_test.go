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
			Client: httpClient{client.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("ListStemcells", func() {
		It("makes a request to list the stemcells", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/stemcell_associations"),
					ghttp.RespondWith(http.StatusOK, `{
				  		"products": [{
							  "guid": "some-guid",
							  "staged_stemcells": [{"os": "ubuntu-trusty", "version": "1234.5"}],
							  "identifier": "some-product",
							  "available_stemcells": [
								{"os": "ubuntu-trusty", "version": "1234.5"},
								{"os": "ubuntu-trusty", "version": "1234.6"}
							  ],
							  "required_stemcells": [{"os": "ubuntu-xenial", "version": "1234.5"}]
						}]
					}`),
				),
			)

			output, err := service.ListMultiStemcells()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID:              "some-guid",
						StagedForDeletion: false,
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
						},
						ProductName: "some-product",
						AvailableVersions: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
							{OS: "ubuntu-trusty", Version: "1234.6"},
						},
						RequiredStemcells: []api.StemcellObject{
							{OS: "ubuntu-xenial", Version: "1234.5"},
						},
					},
				},
			}))
		})

		When("invalid JSON is returned", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_associations"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListMultiStemcells()
				Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of value")))
			})
		})
		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListMultiStemcells()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_associations")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/stemcell_associations"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.ListMultiStemcells()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("AssignMultiStemcells", func() {
		var (
			input api.ProductMultiStemcells
		)

		BeforeEach(func() {
			input = api.ProductMultiStemcells{
				Products: []api.ProductMultiStemcell{
					{
						GUID: "some-guid",
						StagedStemcells: []api.StemcellObject{
							{OS: "ubuntu-trusty", Version: "1234.5"},
							{OS: "ubuntu-trusty", Version: "1234.6"},
						},
					},
				},
			}
		})

		It("makes a request to assign multiple stemcells", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_associations"),
					ghttp.VerifyJSON(`{
				  		"products": [{
						  	"guid": "some-guid",
						  	"staged_stemcells": [
								 {"os": "ubuntu-trusty", "version": "1234.5"},
								 {"os": "ubuntu-trusty", "version": "1234.6"}
						   	]
						}]
            		}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.AssignMultiStemcell(input)
			Expect(err).ToNot(HaveOccurred())
		})

		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

				err := service.AssignMultiStemcell(input)
				Expect(err).To(MatchError(ContainSubstring("could not send api request to PATCH /api/v0/stemcell_associations")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/v0/stemcell_associations"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.AssignMultiStemcell(input)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
