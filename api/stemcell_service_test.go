package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StemcellService", func() {
	Describe("Upload", func() {
		var (
			client           *fakes.HttpClient
			stemcellLocation string
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}

			stemcell, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = stemcell.WriteString("some content")
			Expect(err).NotTo(HaveOccurred())

			err = stemcell.Close()
			Expect(err).NotTo(HaveOccurred())

			stemcellLocation = stemcell.Name()
		})

		AfterEach(func() {
			err := os.Remove(stemcellLocation)
			Expect(err).NotTo(HaveOccurred())
		})

		It("makes a request to upload the stemcell to the OpsManager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			service := api.NewUploadStemcellService(client)

			content, err := os.Open(stemcellLocation)
			Expect(err).NotTo(HaveOccurred())

			output, err := service.Upload(api.StemcellUploadInput{
				ContentLength: 10,
				Stemcell:      content,
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.StemcellUploadOutput{}))

			request := client.DoArgsForCall(0)
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
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewUploadStemcellService(client)

					_, err := service.Upload(api.StemcellUploadInput{})
					Expect(err).To(MatchError("could not make api request to stemcells endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewUploadStemcellService(client)

					_, err := service.Upload(api.StemcellUploadInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
