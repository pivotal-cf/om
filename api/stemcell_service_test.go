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

var _ = Describe("StemcellService", func() {
	Describe("ListStemcells", func() {
		var (
			fakeClient *fakes.HttpClient
			service    api.Api
		)

		BeforeEach(func() {
			fakeClient = &fakes.HttpClient{}
			service = api.New(api.ApiInput{
				Client: fakeClient,
			})
		})

		It("makes a request to list the stemcells", func() {
			fakeClient.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{
                  "products": [
                    {
                      "guid": "some-guid",
                      "staged_stemcell_version": "1234.5",
                      "identifier": "some-product",
                      "available_stemcell_versions": [
                        "1234.5", "1234.6"
                      ]
                    }
                  ]
                }`)),
			}, nil)

			output, err := service.ListStemcells()
			Expect(err).NotTo(HaveOccurred())
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

			request := fakeClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/stemcell_assignments"))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{}, errors.New("some client error"))

					_, err := service.ListStemcells()
					Expect(err).To(MatchError("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_assignments: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)

					_, err := service.ListStemcells()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("AssignStemcells", func() {
		var (
			fakeClient *fakes.HttpClient
			service    api.Api
			input      api.ProductStemcells
		)

		BeforeEach(func() {
			fakeClient = &fakes.HttpClient{}
			service = api.New(api.ApiInput{
				Client: fakeClient,
			})

			input = api.ProductStemcells{
				Products: []api.ProductStemcell{
					{
						GUID:                  "some-guid",
						StagedStemcellVersion: "1234.6",
					},
				},
			}
		})

		It("makes a request to assign the stemcells", func() {
			fakeClient.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := service.AssignStemcell(input)
			Expect(err).NotTo(HaveOccurred())

			request := fakeClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("PATCH"))
			Expect(request.URL.Path).To(Equal("/api/v0/stemcell_assignments"))
			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON(`{
              "products": [
                {
                  "guid": "some-guid",
                  "staged_stemcell_version": "1234.6"
                }
              ]
            }`))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{}, errors.New("some client error"))

					err := service.AssignStemcell(input)
					Expect(err).To(MatchError("could not send api request to PATCH /api/v0/stemcell_assignments: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)

					err := service.AssignStemcell(input)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
