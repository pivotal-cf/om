package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - pivnet versions", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List", func() {
		var (
			pivnetResourceVersion  = "7.8.9"
			pivnetCliVersion       = "2.3.4"
		)

		Context("valid requests", func() {
			It("returns the pivnet component versions", func() {
				response := fmt.Sprintf(`{"pivnet_cli": "%s", "pivnet_resource": "%s"}`, pivnetCliVersion, pivnetResourceVersion)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/versions", apiPrefix)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				versions, err := client.PivnetVersions.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(versions.PivnetCliVersion).To(Equal(pivnetCliVersion))
				Expect(versions.PivnetResourceVersion).To(Equal(pivnetResourceVersion))
			})
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/versions", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.PivnetVersions.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/versions", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.PivnetVersions.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
