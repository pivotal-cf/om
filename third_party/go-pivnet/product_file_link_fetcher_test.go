package pivnet_test

import (
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"
	"net/http"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - ProductFileLinkFetcher", func() {
	var (
		server           *ghttp.Server
		client           pivnet.Client
		pivnetApiAddress string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		pivnetApiAddress = server.URL()

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host: pivnetApiAddress,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewDownloadLink", func() {
		It("returns a url", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("%s/test-endpoint", apiPrefix)),
					ghttp.RespondWith(http.StatusFound, nil,
						http.Header{
							"Location": []string{"http://example.com"},
						},
					),
				),
			)

			linkFetcher := pivnet.NewProductFileLinkFetcher("/test-endpoint", client)
			link, err := linkFetcher.NewDownloadLink()
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("http://example.com"))
		})
	})
})
