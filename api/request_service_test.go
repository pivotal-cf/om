package api_test

import (
	"net/http"
	"strings"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("RequestService", func() {
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

	It("makes a request against the api and returns a response", func() {
		client.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v0/api/endpoint"),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyBody([]byte("some-request-body")),
				ghttp.RespondWith(http.StatusTeapot, "", map[string][]string{"Content-Type": {"application/json"}}),
			),
		)

		output, err := service.Curl(api.RequestServiceCurlInput{
			Method:  "PUT",
			Path:    "/api/v0/api/endpoint",
			Data:    strings.NewReader("some-request-body"),
			Headers: http.Header{"Content-Type": []string{"application/json"}},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(output.StatusCode).To(Equal(http.StatusTeapot))
		Expect(output.Headers.Get("Content-Type")).To(Equal("application/json"))
	})

	When("the request cannot be constructed", func() {
		It("returns an error", func() {
			_, err := service.Curl(api.RequestServiceCurlInput{
				Method: "PUT",
				Path:   "%%%",
				Data:   strings.NewReader("some-request-body"),
			})

			Expect(err).To(MatchError(ContainSubstring("failed constructing request:")))
		})
	})

	When("the request cannot be made", func() {
		It("returns an error", func() {
			client.Close()

			_, err := service.Curl(api.RequestServiceCurlInput{
				Method: "PUT",
				Path:   "/api/v0/api/endpoint",
				Data:   strings.NewReader("some-request-body"),
			})

			Expect(err).To(MatchError(ContainSubstring("failed submitting request:")))
		})
	})
})
