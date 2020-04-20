package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"net/http"
)

var _ = Describe("ConfigureOpsmanService", func() {
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

	Describe("ListDeployedProductCredentials", func() {
		It("lists credential references", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/settings/pivotal_network_settings"),
					ghttp.RespondWith(http.StatusOK, `{
					  "success": true
					}`),
					ghttp.VerifyJSON("{ \"pivotal_network_settings\": { \"api_token\": \"some-api-token\" }}"),
				),
			)

			err := service.UpdatePivnetToken("some-api-token")
			Expect(err).ToNot(HaveOccurred())
		})

		When("the api returns an error", func() {
			It("lists credential references", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/settings/pivotal_network_settings"),
						ghttp.RespondWith(http.StatusInternalServerError, "{}"),
					),
				)

				err := service.UpdatePivnetToken("some-api-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("500 Internal Server Error"))
			})
		})
	})
})
