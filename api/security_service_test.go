package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("Security", func() {
	Describe("Fetch Root CA Cert", func() {
		var (
			client  *fakes.HttpClient
			service api.Api
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			service = api.New(api.ApiInput{
				Client: client,
			})
		})

		It("gets the root CA cert", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"root_ca_certificate_pem": "some-response-cert"}`)),
			}, nil)

			output, err := service.GetSecurityRootCACertificate()
			Expect(err).NotTo(HaveOccurred())

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/security/root_ca_certificate"))
			Expect(output).To(Equal("some-response-cert"))
		})

		Context("error cases", func() {
			It("returns error if request fails to submit", func() {
				client.DoReturns(&http.Response{}, errors.New("some-error"))

				_, err := service.GetSecurityRootCACertificate()
				Expect(err).To(MatchError("failed to submit request: could not send api request to GET /api/v0/security/root_ca_certificate: some-error"))
			})

			It("returns error when response contains non-200 status code", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)

				_, err := service.GetSecurityRootCACertificate()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})

			It("returns error if response fails to unmarshal", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`%%%`)),
				}, nil)

				_, err := service.GetSecurityRootCACertificate()
				Expect(err).To(MatchError(ContainSubstring("failed to unmarshal response: invalid character")))
			})
		})
	})
})
