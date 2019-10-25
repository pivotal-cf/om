package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("DisableDirectorVerifiersService", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()
		service = api.New(api.ApiInput{
			Client: httpClient{serverURI: client.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("ListDirectorVerifiers", func() {
		It("lists available director verifiers", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/verifiers/install_time"),
					ghttp.RespondWith(http.StatusOK, `{
						"verifiers": [{
							"type": "some-verifier-type",
							"enabled": true
						}, {
							"type": "another-verifier-type",
							"enabled": false
						}]
					}`),
				),
			)

			output, err := service.ListDirectorVerifiers()
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(Equal([]api.Verifier{
				{
					Type:    "some-verifier-type",
					Enabled: true,
				}, {
					Type:    "another-verifier-type",
					Enabled: false,
				},
			}))
		})

		It("returns an error when not 200-OK", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/verifiers/install_time"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			_, err := service.ListDirectorVerifiers()
			Expect(err).To(MatchError(ContainSubstring("unexpected response")))
		})

		It("returns an error when the http request could not be made", func() {
			client.Close()

			_, err := service.ListDirectorVerifiers()
			Expect(err).To(MatchError(ContainSubstring("could not make api request to list_director_verifiers endpoint")))
		})

		It("returns an error when the response is not JSON", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/director/verifiers/install_time"),
					ghttp.RespondWith(http.StatusOK, `invalid JSON`),
				),
			)

			_, err := service.ListDirectorVerifiers()
			Expect(err).To(MatchError(ContainSubstring("could not unmarshal list_director_verifiers response")))
		})
	})

	Describe("DisableDirectorVerifiers", func() {
		It("disables a list of director verifiers", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/director/verifiers/install_time/some-verifier-type"),
					ghttp.VerifyJSON(`{"enabled": false}`),
					ghttp.RespondWith(http.StatusOK, `{"type":"some-verifier-type", "enabled":false}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/director/verifiers/install_time/another-verifier-type"),
					ghttp.VerifyJSON(`{"enabled": false}`),
					ghttp.RespondWith(http.StatusOK, `{"type":"another-verifier-type", "enabled":false}`),
				),
			)

			err := service.DisableDirectorVerifiers([]string{"some-verifier-type", "another-verifier-type"})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the endpoint returns a non-200-OK status code", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/director/verifiers/install_time/some-verifier-type"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err := service.DisableDirectorVerifiers([]string{"some-verifier-type"})
				Expect(err).To(MatchError(ContainSubstring("unexpected response")))
			})

			It("returns an error when the http request could not be made", func() {
				client.Close()

				err := service.DisableDirectorVerifiers([]string{"some-verifier-type"})
				Expect(err).To(MatchError(ContainSubstring("could not make api request to disable_director_verifiers endpoint")))
			})
		})
	})
})
