package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("Credentials", func() {
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

	Describe("GetDeployedDirectorCredential", func() {
		It("fetch a credential reference", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials/bosh_commandline_credentials"),
					ghttp.RespondWith(http.StatusOK, `{
						"credential": "BOSH_CLIENT=ops_manager BOSH_CLIENT_SECRET=foo BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate BOSH_ENVIRONMENT=10.0.0.10 bosh"
					}`),
				),
			)

			output, err := service.GetBoshEnvironment()
			Expect(err).NotTo(HaveOccurred())

			Expect(output.Client).To(Equal("ops_manager"))
			Expect(output.ClientSecret).To(Equal("foo"))
			Expect(output.Environment).To(Equal("10.0.0.10"))
		})

		When("the client can't connect to the server", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetBoshEnvironment()
				Expect(err).To(MatchError(ContainSubstring("could not make api request")))
			})
		})

		When("the server won't fetch credential references", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials/bosh_commandline_credentials"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				_, err := service.GetBoshEnvironment()
				Expect(err).To(MatchError(ContainSubstring("request failed")))
			})
		})

		When("the response is not JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials/bosh_commandline_credentials"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetBoshEnvironment()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
			})
		})
	})
})
