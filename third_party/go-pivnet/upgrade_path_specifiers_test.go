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

var _ = Describe("PivnetClient - upgrade path specifiers", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService

		productSlug string
		releaseID   int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		productSlug = "some-product"
		releaseID = 2345

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
		It("returns the upgrade path specifiers", func() {
			response := pivnet.UpgradePathSpecifiersResponse{
				UpgradePathSpecifiers: []pivnet.UpgradePathSpecifier{
					{
						ID:        9876,
						Specifier: "1.2.*",
					},
					{
						ID:        8765,
						Specifier: "~>2.3.4",
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_path_specifiers",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			upgradePathSpecifiers, err := client.UpgradePathSpecifiers.List(productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(upgradePathSpecifiers).To(HaveLen(2))
			Expect(upgradePathSpecifiers[0].ID).To(Equal(9876))
			Expect(upgradePathSpecifiers[1].ID).To(Equal(8765))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.UpgradePathSpecifiers.List(productSlug, releaseID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UpgradePathSpecifiers.List(productSlug, releaseID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get", func() {
		var (
			upgradePathSpecifierID int
		)

		BeforeEach(func() {
			upgradePathSpecifierID = 1234
		})

		It("returns the upgrade path specifier", func() {
			response := pivnet.UpgradePathSpecifierResponse{
				UpgradePathSpecifier: pivnet.UpgradePathSpecifier{
					ID:        upgradePathSpecifierID,
					Specifier: "1.2.*",
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
						apiPrefix,
						productSlug,
						releaseID,
						upgradePathSpecifierID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			upgradePathSpecifier, err := client.UpgradePathSpecifiers.Get(
				productSlug,
				releaseID,
				upgradePathSpecifierID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(upgradePathSpecifier.ID).To(Equal(upgradePathSpecifierID))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							upgradePathSpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.UpgradePathSpecifiers.Get(
					productSlug,
					releaseID,
					upgradePathSpecifierID,
				)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							upgradePathSpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UpgradePathSpecifiers.Get(
					productSlug,
					releaseID,
					upgradePathSpecifierID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Create", func() {
		var (
			specifier string
		)

		BeforeEach(func() {
			specifier = "1.5.*"
		})

		It("creates the upgrade path specifier", func() {
			expectedRequestBody := fmt.Sprintf(
				`{"upgrade_path_specifier":{"specifier":"%s"}}`,
				specifier,
			)

			response := pivnet.UpgradePathSpecifierResponse{
				UpgradePathSpecifier: pivnet.UpgradePathSpecifier{
					ID:        1234,
					Specifier: specifier,
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_path_specifiers",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, response),
				),
			)

			upgradePathSpecifier, err := client.UpgradePathSpecifiers.Create(
				productSlug,
				releaseID,
				specifier,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(upgradePathSpecifier.ID).To(Equal(1234))
			Expect(upgradePathSpecifier.Specifier).To(Equal(specifier))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.UpgradePathSpecifiers.Create(
					productSlug,
					releaseID,
					specifier,
				)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.UpgradePathSpecifiers.Create(
					productSlug,
					releaseID,
					specifier,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete", func() {
		var (
			upgradePathSpecifierID int
		)

		BeforeEach(func() {
			upgradePathSpecifierID = 1234
		})

		It("deletes the upgrade path specifier", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
						apiPrefix,
						productSlug,
						releaseID,
						upgradePathSpecifierID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, nil),
				),
			)

			err := client.UpgradePathSpecifiers.Delete(
				productSlug,
				releaseID,
				upgradePathSpecifierID,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							upgradePathSpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				err := client.UpgradePathSpecifiers.Delete(
					productSlug,
					releaseID,
					upgradePathSpecifierID,
				)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_path_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							upgradePathSpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.UpgradePathSpecifiers.Delete(
					productSlug,
					releaseID,
					upgradePathSpecifierID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
