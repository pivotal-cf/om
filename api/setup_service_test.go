package api_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Setup", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			UnauthedClient: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("Setup", func() {
		It("makes a request to setup the OpsManager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/setup"),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWith(http.StatusOK, `{
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
					}`),
				),
			)

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
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.SetupOutput{}))
		})

		When("the client fails to make the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.Setup(api.SetupInput{})
				Expect(err).To(MatchError(ContainSubstring("could not make api request to setup endpoint")))
			})
		})

		When("the api returns an unexpected status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/setup"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"error": "something bad happened"}`),
					),
				)

				_, err := service.Setup(api.SetupInput{})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("EnsureAvailability", func() {
		When("the availability endpoint returns an unexpected status code", func() {
			It("returns a helpful error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusTeapot, `{"error": "something bad happened"}`),
					),
				)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Unexpected response code: 418 I'm a teapot"))
			})
		})

		When("the availability endpoint returns an OK status with an unexpected body", func() {
			It("returns a helpful error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusOK, "some body"),
					),
				)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Received OK with an unexpected body: some body"))
			})
		})

		When("the availability endpoint returns a found status with an unexpected location header", func() {
			It("returns a helpful error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"https://some-opsman/something/else"}}),
					),
				)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError("Unexpected redirect location: /something/else"))
			})
		})

		When("the authentication mechanism has not been setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {fmt.Sprintf("%s/setup", client.URL())}}),
					),
				)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusUnstarted,
				}))
			})
		})

		When("the authentication mechanism is currently being setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusOK, "Waiting for authentication system to start..."),
					),
				)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusPending,
				}))
			})
		})

		When("the authentication mechanism is completely setup", func() {
			It("makes a request to determine the availability of the OpsManager authentication mechanism", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"https://some-opsman/auth/cloudfoundry"}}),
					),
				)

				output, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(api.EnsureAvailabilityOutput{
					Status: api.EnsureAvailabilityStatusComplete,
				}))
			})
		})

		When("the request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError(ContainSubstring("could not make request round trip")))
			})
		})

		When("the location header cannot be parsed", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/login/ensure_availability"),
						ghttp.RespondWith(http.StatusFound, "", map[string][]string{"Location": {"%%%%%%"}}),
					),
				)

				_, err := service.EnsureAvailability(api.EnsureAvailabilityInput{})
				Expect(err).To(MatchError(ContainSubstring(`parse "%%%%%%": invalid URL escape "%%%"`)))
			})
		})
	})
})
