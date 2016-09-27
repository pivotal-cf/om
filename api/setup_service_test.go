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

var _ = Describe("SetupService", func() {
	Describe("Setup", func() {
		It("makes a request to setup the OpsManager", func() {
			client := &fakes.Client{}
			client.DoCall.Returns.Response = &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}

			service := api.NewSetupService(client)

			output, err := service.Setup(api.SetupInput{
				IdentityProvider:                 "some-provider",
				AdminUserName:                    "some-username",
				AdminPassword:                    "some-password",
				AdminPasswordConfirmation:        "some-password-confirmation",
				DecryptionPassphrase:             "some-passphrase",
				DecryptionPassphraseConfirmation: "some-passphrase-confirmation",
				EULAAccepted:                     true,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.SetupOutput{}))

			Expect(client.DoCall.Receives.Request).NotTo(BeNil())
			Expect(client.DoCall.Receives.Request.Method).To(Equal("POST"))
			Expect(client.DoCall.Receives.Request.URL.Path).To(Equal("/api/v0/setup"))
			Expect(client.DoCall.Receives.Request.Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := ioutil.ReadAll(client.DoCall.Receives.Request.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(MatchJSON(`{
				"setup": {
    			"identity_provider": "some-provider",
					"admin_user_name": "some-username",
					"admin_password": "some-password",
					"admin_password_confirmation": "some-password-confirmation",
					"decryption_passphrase": "some-passphrase",
					"decryption_passphrase_confirmation":"some-passphrase-confirmation",
					"eula_accepted": "true"
				}
			}`))
		})

		Context("failure cases", func() {
			Context("when the client fails to make the request", func() {
				It("returns an error", func() {
					client := &fakes.Client{}
					client.DoCall.Returns.Error = errors.New("could not make request")

					service := api.NewSetupService(client)

					_, err := service.Setup(api.SetupInput{})
					Expect(err).To(MatchError("could not make api request to setup endpoint: could not make request"))
				})
			})

			Context("when the api returns an unexpected status code", func() {
				It("returns an error", func() {
					client := &fakes.Client{}
					client.DoCall.Returns.Response = &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
						Body:       ioutil.NopCloser(strings.NewReader(`{"error": "something bad happened"}`)),
					}

					service := api.NewSetupService(client)

					_, err := service.Setup(api.SetupInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("EnsureAvailability", func() {
		Context("when the authentication mechanism has not been setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client := &fakes.Client{}
				client.RoundTripCall.Returns.Response = &http.Response{
					StatusCode: http.StatusFound,
					Header: http.Header{
						"Location": []string{"/login/setup"},
					},
				}

				service := api.NewSetupService(client)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}))
			})
		})

		Context("when the authentication mechanism is currently being setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client := &fakes.Client{}
				client.RoundTripCall.Returns.Response = &http.Response{
					StatusCode: http.StatusOK,
				}

				service := api.NewSetupService(client)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusPending,
				}))
			})
		})

		Context("when the authentication mechanism is completely setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client := &fakes.Client{}
				client.RoundTripCall.Returns.Response = &http.Response{
					StatusCode: http.StatusFound,
					Header: http.Header{
						"Location": []string{"/auth/cloudfoundry"},
					},
				}

				service := api.NewSetupService(client)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}))
			})
		})
	})
})
