package api_test

import (
	"errors"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisableDirectorVerifiersService", func() {
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

	Describe("ListDirectorVerifiers", func() {
		It("lists available director verifiers", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"verifiers": [
							{
								"type": "some-verifier-type",
								"enabled": true
							},
							{
								"type": "another-verifier-type",
								"enabled": false
							}
						]
					}`)),
				}, nil
			}

			output, err := service.ListDirectorVerifiers()
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(Equal([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: true,
				},
				{
					Type:    "another-verifier-type",
					Enabled: false,
				},
			}))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/director/verifiers/install_time"))
		})

		Context("failure cases", func() {
			It("returns an error when not 200-OK", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusInternalServerError,
						Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}

				_, err := service.ListDirectorVerifiers()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unexpected response"))
			})

			It("returns an error when the http request could not be made", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("something happened")
				}

				_, err := service.ListDirectorVerifiers()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not make api request to list_director_verifiers endpoint"))
			})

			It("returns an error when the response is not JSON", func() {
				client.DoStub = func(req *http.Request) (response *http.Response, e error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`invalid JSON`))}, nil
				}

				_, err := service.ListDirectorVerifiers()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not unmarshal list_director_verifiers response"))
			})
		})
	})

	Describe("DisableDirectorVerifiers", func() {
		It("disables a list of director verifiers", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/staged/director/verifiers/install_time/some-verifier-type" {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{"type":"some-verifier-type", "enabled":false}`))}, nil
				} else {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{"type":"another-verifier-type", "enabled":false}`))}, nil
				}
			}

			err := service.DisableDirectorVerifiers([]string{"some-verifier-type", "another-verifier-type"})
			Expect(err).NotTo(HaveOccurred())

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal(http.MethodPut))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/director/verifiers/install_time/some-verifier-type"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).ToNot(HaveOccurred())
			defer request.Body.Close()

			Expect(string(body)).To(Equal(`{ "enabled": false }`))

			request = client.DoArgsForCall(1)
			Expect(request.Method).To(Equal("PUT"))
			Expect(request.URL.Path).To(Equal("/api/v0/staged/director/verifiers/install_time/another-verifier-type"))
		})

		Context("failure cases", func() {
			It("returns an error when the endpoint returns a non-200-OK status code", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusInternalServerError,
						Body: ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}

				err := service.DisableDirectorVerifiers([]string{"some-verifier-type"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unexpected response"))
			})

			It("returns an error when the http request could not be made", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("something happened")
				}

				err := service.DisableDirectorVerifiers([]string{"some-verifier-type"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not make api request to disable_director_verifiers endpoint"))
			})
		})
	})
})
