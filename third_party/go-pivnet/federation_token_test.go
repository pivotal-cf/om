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

var _ = Describe("PivnetClient - federation token", func() {
	type requestBody struct {
		ProductID string `json:"product_id"`
	}

	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		mockedResponse         interface{}
		responseStatusCode     int
		expectedRequestBody    requestBody
		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService

		productSlug             string
		expectedFederationToken pivnet.FederationToken
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

	JustBeforeEach(func() {
		expectedRequestBody = requestBody{
			ProductID: productSlug,
		}

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf("%s/federation_token", apiPrefix),
				),
				ghttp.VerifyJSONRepresenting(&expectedRequestBody),
				ghttp.RespondWithJSONEncoded(responseStatusCode, mockedResponse),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Generate a federated token", func() {
		BeforeEach(func() {
			productSlug = "banana"

			mockedResponse = pivnet.FederationToken{
				AccessKeyID:     "some-AccessKeyID",
				SecretAccessKey: "some-SecretAccessKey",
				SessionToken:    "some-SessionToken",
				Bucket:          "some-bucket",
				Region:          "some-region",
			}

			responseStatusCode = http.StatusOK
			expectedFederationToken = pivnet.FederationToken{
				AccessKeyID:     "some-AccessKeyID",
				SecretAccessKey: "some-SecretAccessKey",
				SessionToken:    "some-SessionToken",
				Bucket:          "some-bucket",
				Region:          "some-region",
			}
		})

		It("returns the federated token without error", func() {
			federationToken, err := client.FederationToken.GenerateFederationToken(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(federationToken).ToNot(Equal(nil))
			Expect(federationToken).To(Equal(expectedFederationToken))
		})
	})

	Describe("Err when trying to generate token for restricted product", func() {
		BeforeEach(func() {
			productSlug = "something-i-dont-manage"

			mockedResponse = pivnetErr{Message: "only available for product admins and partner product admins"}

			responseStatusCode = http.StatusForbidden
		})

		It("returns a 403 error", func() {
			federationToken, err := client.FederationToken.GenerateFederationToken(
				productSlug,
			)

			Expect(federationToken).To(Equal(pivnet.FederationToken{}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403 - only available for product admins and partner product admins"))
		})
	})
})
