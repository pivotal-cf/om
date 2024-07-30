package api_test

import (
	"net/http"
	"strings"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("UploadStemcellService", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			ProgressClient: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("UploadStemcell", func() {
		It("makes a request to upload the stemcell to the OpsManager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
					ghttp.VerifyContentType("some content-type"),
					ghttp.VerifyBody([]byte("some content")),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Expect(req.ContentLength).To(Equal(int64(12)))

						_, err := w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)

			output, err := service.UploadStemcell(api.StemcellUploadInput{
				ContentLength: 12,
				Stemcell:      strings.NewReader("some content"),
				ContentType:   "some content-type",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.StemcellUploadOutput{}))
		})

		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.UploadStemcell(api.StemcellUploadInput{})
				Expect(err).To(MatchError(ContainSubstring("could not make api request to stemcells endpoint")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/stemcells"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.UploadStemcell(api.StemcellUploadInput{})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})
})
