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
	Describe("UploadStemcell", func() {
		var (
			progressClient *fakes.HttpClient
		)

		BeforeEach(func() {
			progressClient = &fakes.HttpClient{}
		})

		It("makes a request to upload the stemcell to the OpsManager", func() {
			progressClient.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			service := api.NewUploadStemcellService(progressClient)

			output, err := service.UploadStemcell(api.StemcellUploadInput{
				ContentLength: 10,
				Stemcell:      strings.NewReader("some content"),
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.StemcellUploadOutput{}))

			request := progressClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/stemcells"))
			Expect(request.ContentLength).To(Equal(int64(10)))
			Expect(request.Header.Get("Content-Type")).To(Equal("some content-type"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal("some content"))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewUploadStemcellService(progressClient)

					_, err := service.UploadStemcell(api.StemcellUploadInput{})
					Expect(err).To(MatchError("could not make api request to stemcells endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewUploadStemcellService(progressClient)

					_, err := service.UploadStemcell(api.StemcellUploadInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
