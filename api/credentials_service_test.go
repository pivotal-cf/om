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

var _ = Describe("CredentialsService", func() {
	var (
		client  *fakes.HttpClient
		bar     *fakes.Progress
		service api.CredentialsService
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		bar = &fakes.Progress{}

		service = api.NewCredentialsService(client, bar)
	})

	Describe("ListCredentials", func() {

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
			output, err := service.ListCredentials("some-deployed-product-guid")
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
					_, err := service.ListCredentials("invalid-product")
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			Context("when the server won't fetch credential references", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.ListCredentials("")
					Expect(err).To(MatchError(ContainSubstring("request failed")))
				})
			})

			Context("when the response is not JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`asdf`)),
					}, nil)

					_, err := service.ListCredentials("some-deployed-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
				})
			})
		})
	})

	Describe("FetchCredential", func() {

		It("fetch a credential reference", func() {
			var path string

			client.DoStub = func(req *http.Request) (*http.Response, error) {
				path = req.URL.Path

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(
						strings.NewReader(`{"credential":{"type":"rsa_cert_credentials", "credential": true, "value":{"private_key_pem":"some-private-key", "cert_pem":"some-cert-pem"}}}`),
					),
				}, nil
			}
			output, err := service.FetchCredential("some-deployed-product-guid", ".properties.some-credentials")
			Expect(err).NotTo(HaveOccurred())

			Expect(path).To(Equal("/api/v0/deployed/products/some-deployed-product-guid/credentials/.properties.some-credentials"))
			Expect(output.Credential.Value["private_key_pem"]).To(Equal("some-private-key"))
			Expect(output.Credential.Value["cert_pem"]).To(Equal("some-cert-pem"))
		})

		Describe("errors", func() {
			Context("the client can't connect to the server", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))
					_, err := service.FetchCredential("invalid-product", "")
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			Context("when the server won't fetch credential references", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.FetchCredential("invalid-product", "")
					Expect(err).To(MatchError(ContainSubstring("request failed")))
				})
			})

			Context("when the response is not JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`asdf`)),
					}, nil)

					_, err := service.FetchCredential("some-deployed-product-guid", "")
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
				})
			})
		})
	})
})
