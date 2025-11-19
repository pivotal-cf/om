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

var _ = Describe("PivnetClient - EULA", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
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
		It("returns all EULAs", func() {
			response := `{"eulas": [{"id":1,"name":"eula1"},{"id": 2,"name":"eula2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			eulas, err := client.EULA.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(eulas).To(HaveLen(2))

			Expect(eulas[0].ID).To(Equal(1))
			Expect(eulas[0].Name).To(Equal("eula1"))
			Expect(eulas[1].ID).To(Equal(2))
			Expect(eulas[1].Name).To(Equal("eula2"))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.EULA.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
						ghttp.RespondWith(http.StatusOK, "%%%"),
					),
				)

				_, err := client.EULA.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get", func() {
		var (
			eulaSlug string
		)

		BeforeEach(func() {
			eulaSlug = "eula_1"
		})

		It("returns the EULA for the provided eula slug", func() {
			response := `{"id":1,"name":"eula1","slug":"eula_1"}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas/%s", apiPrefix, eulaSlug)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			eula, err := client.EULA.Get(eulaSlug)
			Expect(err).NotTo(HaveOccurred())

			Expect(eula.ID).To(Equal(1))
			Expect(eula.Name).To(Equal("eula1"))
			Expect(eula.Slug).To(Equal(eulaSlug))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas/%s", apiPrefix, eulaSlug)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.EULA.Get(eulaSlug)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas/%s", apiPrefix, eulaSlug)),
						ghttp.RespondWith(http.StatusOK, "%%%"),
					),
				)

				_, err := client.EULA.Get(eulaSlug)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Accept", func() {
		var (
			releaseID         int
			productSlug       string
			EULAAcceptanceURL string
		)

		BeforeEach(func() {
			productSlug = "banana-slug"
			releaseID = 42
			EULAAcceptanceURL = fmt.Sprintf(apiPrefix+"/products/%s/releases/%d/pivnet_resource_eula_acceptance", productSlug, releaseID)
			fakeAccessTokenService.AccessTokenReturns(token, nil)
		})

		It("accepts the EULA for a given release and product ID", func() {
			response := fmt.Sprintf(`{"accepted_at": "2016-01-11","_links":{}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", EULAAcceptanceURL),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyJSON(`{}`),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			Expect(client.EULA.Accept(productSlug, releaseID)).To(Succeed())
		})

		Context("when any other non-200 status code comes back", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", EULAAcceptanceURL),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
						ghttp.VerifyJSON(`{}`),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				err := client.EULA.Accept(productSlug, releaseID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})
})
