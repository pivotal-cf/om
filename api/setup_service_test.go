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

var _ = Describe("Setup", func() {
	var (
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			UnauthedClient: client,
		})
	})

	Describe("Setup", func() {
		It("makes a request to setup the OpsManager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			output, err := service.Setup(api.SetupInput{
				IdentityProvider:                 "some-provider",
				AdminUserName:                    "some-username",
				AdminPassword:                    "some-password",
				AdminPasswordConfirmation:        "some-password-confirmation",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase-confirmation",
				EULAAccepted:                     "true",
				HTTPProxyURL:                     "http://http-proxy.com",
				HTTPSProxyURL:                    "http://https-proxy.com",
				NoProxy:                          "10.10.10.10,11.11.11.11",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.SetupOutput{}))

			request := client.DoArgsForCall(0)
			Expect(request).NotTo(BeNil())
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/setup"))
			Expect(request.Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON(`{
				"setup": {
    			"identity_provider": "some-provider",
					"admin_user_name": "some-username",
					"admin_password": "some-password",
					"admin_password_confirmation": "some-password-confirmation",
					"decryption_passphrase": "some-passphrase",
					"decryption_passphrase_confirmation":"some-passphrase-confirmation",
					"eula_accepted": "true",
					"http_proxy": "http://http-proxy.com",
					"https_proxy": "http://https-proxy.com",
					"no_proxy": "10.10.10.10,11.11.11.11"
				}
			}`))
		})

		Context("failure cases", func() {
			Context("when the client fails to make the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("could not make request"))

					_, err := service.Setup(api.SetupInput{})
					Expect(err).To(MatchError("could not make api request to setup endpoint: could not send api request to POST /api/v0/setup: could not make request"))
				})
			})

			Context("when the api returns an unexpected status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
						Body:       ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened"}`)),
					}, nil)

					_, err := service.Setup(api.SetupInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("EnsureAvailability", func() {
		Context("when the availability endpoint returns an unexpected status code", func() {
			It("returns a helpful error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Unexpected response code: 418 I'm a teapot"))
			})
		})

		Context("when the availability endpoint returns an OK status with an unexpected body", func() {
			It("returns a helpful error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("some body")),
				}, nil)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Received OK with an unexpected body: some body"))
			})
		})

		Context("when the availability endpoint returns a found status with an unexpected location header", func() {
			It("returns a helpful error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusFound,
					Header: http.Header{
						"Location": []string{"https://some-opsman/something/else"},
					},
					Body: ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Unexpected redirect location: /something/else"))
			})
		})

		Context("when the authentication mechanism has not been setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusFound,
					Header: http.Header{
						"Location": []string{"https://some-opsman/setup"},
					},
					Body: ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}))
			})
		})

		Context("when the authentication mechanism is currently being setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("Waiting for authentication system to start...")),
				}, nil)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusPending,
				}))
			})
		})

		Context("when the authentication mechanism is completely setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusFound,
					Header: http.Header{
						"Location": []string{"https://some-opsman/auth/cloudfoundry"},
					},
					Body: ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the request fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("failed to make round trip"))

					_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
					Expect(err).To(MatchError("could not make request round trip: failed to make round trip"))
				})
			})

			Context("when the location header cannot be parsed", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusFound,
						Header: http.Header{
							"Location": []string{"%%%%%%"},
						},
						Body: ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
					Expect(err).To(MatchError("could not parse redirect url: parse %%%%%%: invalid URL escape \"%%%\""))
				})
			})
		})
	})
})
