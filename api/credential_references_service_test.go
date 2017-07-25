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

var _ = Describe("CredentialReferencesService", func() {
	var (
		client  *fakes.HttpClient
		bar     *fakes.Progress
		service api.CredentialReferencesService
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		bar = &fakes.Progress{}

		service = api.NewCredentialReferencesService(client, bar)
	})

	Describe("List", func() {

		It("lists credential references", func() {
			var path string

			client.DoStub = func(req *http.Request) (*http.Response, error) {
				path = req.URL.Path

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(
						strings.NewReader(`{"credentials":[".properties.some-credentials",".my-job.some-credentials"]}`),
					),
				}, nil
			}
			output, err := service.List("some-deployed-product-guid")
			Expect(err).NotTo(HaveOccurred())

			Expect(path).To(Equal("/api/v0/deployed/products/some-deployed-product-guid/credentials"))
			Expect(output.Credentials).To(ConsistOf(
				[]string{
					".properties.some-credentials",
					".my-job.some-credentials",
				},
			))
		})

		Describe("errors", func() {
			Context("the client can't connect to the server", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))
					_, err := service.List("invalid-product")
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			Context("when the server won't fetch credential references", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.List("")
					Expect(err).To(MatchError(ContainSubstring("request failed")))
				})
			})

			Context("when the response is not JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`asdf`)),
					}, nil)

					_, err := service.List("some-deployed-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
				})
			})
		})
	})
})
