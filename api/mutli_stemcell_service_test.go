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
                      "staged_stemcells": [
						{"os": "ubuntu-trusty", "version": "1234.5"}
                      ],
                      "identifier": "some-product",
                      "available_stemcells": [
                        {"os": "ubuntu-trusty", "version": "1234.5"},
                        {"os": "ubuntu-trusty", "version": "1234.6"}
                      ],
                      "required_stemcells": [ {"os": "ubuntu-xenial", "version": "1234.5"} ]
                    }
                  ]
                }`)),
			}, nil)

			output, err := service.ListMultiStemcells()
			Expect(err).NotTo(HaveOccurred())
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

			request := fakeClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/stemcell_associations"))
		})

		Context("when an error occurs", func() {
			When("invalid JSON is returned", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{invalidJSON}")),
					}, nil)

					_, err := service.ListMultiStemcells()
					Expect(err).To(MatchError(ContainSubstring("invalid JSON: invalid character 'i' looking for beginning of object key string")))
				})
			})
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{}, errors.New("some client error"))

					_, err := service.ListMultiStemcells()
					Expect(err).To(MatchError("could not make api request to list stemcells: could not send api request to GET /api/v0/stemcell_associations: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)

					_, err := service.ListMultiStemcells()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("AssignMultiStemcells", func() {
		var (
			fakeClient *fakes.HttpClient
			service    api.Api
			input      api.ProductMultiStemcells
		)

		BeforeEach(func() {
			fakeClient = &fakes.HttpClient{}
			service = api.New(api.ApiInput{
				Client: fakeClient,
			})

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
			fakeClient.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := service.AssignMultiStemcell(input)
			Expect(err).NotTo(HaveOccurred())

			request := fakeClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("PATCH"))
			Expect(request.URL.Path).To(Equal("/api/v0/stemcell_associations"))
			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON(`{
              "products": [
                {
                  "guid": "some-guid",
                  "staged_stemcells": [
					 {"os": "ubuntu-trusty", "version": "1234.5"},
					 {"os": "ubuntu-trusty", "version": "1234.6"}
                   ]
                }
              ]
            }`))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{}, errors.New("some client error"))

					err := service.AssignMultiStemcell(input)
					Expect(err).To(MatchError("could not send api request to PATCH /api/v0/stemcell_associations: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					fakeClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)

					err := service.AssignMultiStemcell(input)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
