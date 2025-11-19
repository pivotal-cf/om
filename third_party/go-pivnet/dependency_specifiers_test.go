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

var _ = Describe("PivnetClient - dependency specifiers", func() {
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
		It("returns the dependency specifiers", func() {
			response := pivnet.DependencySpecifiersResponse{
				DependencySpecifiers: []pivnet.DependencySpecifier{
					{
						ID:        9876,
						Specifier: "1.2.*",
						Product: pivnet.Product{
							ID:   23,
							Name: "Product 23",
						},
					},
					{
						ID:        8765,
						Specifier: "2.3.*",
						Product: pivnet.Product{
							ID:   23,
							Name: "Product 23",
						},
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependency_specifiers",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			dependencySpecifiers, err := client.DependencySpecifiers.List(productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencySpecifiers).To(HaveLen(2))
			Expect(dependencySpecifiers[0].ID).To(Equal(9876))
			Expect(dependencySpecifiers[1].ID).To(Equal(8765))
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
							"%s/products/%s/releases/%d/dependency_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.DependencySpecifiers.List(productSlug, releaseID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/dependency_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.DependencySpecifiers.List(productSlug, releaseID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get", func() {
		var (
			dependencySpecifierID int
		)

		BeforeEach(func() {
			dependencySpecifierID = 1234
		})

		It("returns the dependency specifier", func() {
			response := pivnet.DependencySpecifierResponse{
				DependencySpecifier: pivnet.DependencySpecifier{
					ID:        dependencySpecifierID,
					Specifier: "1.2.*",
					Product: pivnet.Product{
						ID:   23,
						Name: "Product 23",
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependency_specifiers/%d",
						apiPrefix,
						productSlug,
						releaseID,
						dependencySpecifierID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			dependencySpecifier, err := client.DependencySpecifiers.Get(
				productSlug,
				releaseID,
				dependencySpecifierID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencySpecifier.ID).To(Equal(dependencySpecifierID))
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
							"%s/products/%s/releases/%d/dependency_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							dependencySpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.DependencySpecifiers.Get(
					productSlug,
					releaseID,
					dependencySpecifierID,
				)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/dependency_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							dependencySpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.DependencySpecifiers.Get(
					productSlug,
					releaseID,
					dependencySpecifierID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Create", func() {
		var (
			dependentProductSlug string
			specifier            string
		)

		BeforeEach(func() {
			dependentProductSlug = "other-product"
			specifier = "1.5.*"
		})

		It("creates the dependency specifier", func() {
			expectedRequestBody := fmt.Sprintf(
				`{"dependency_specifier":{"product_slug":"%s","specifier":"%s"}}`,
				dependentProductSlug,
				specifier,
			)

			response := pivnet.DependencySpecifierResponse{
				DependencySpecifier: pivnet.DependencySpecifier{
					ID:        1234,
					Specifier: specifier,
					Product: pivnet.Product{
						ID:   23,
						Name: dependentProductSlug,
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependency_specifiers",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, response),
				),
			)

			dependencySpecifier, err := client.DependencySpecifiers.Create(
				productSlug,
				releaseID,
				dependentProductSlug,
				specifier,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(dependencySpecifier.ID).To(Equal(1234))
			Expect(dependencySpecifier.Specifier).To(Equal(specifier))
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
							"%s/products/%s/releases/%d/dependency_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.DependencySpecifiers.Create(
					productSlug,
					releaseID,
					dependentProductSlug,
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
							"%s/products/%s/releases/%d/dependency_specifiers",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.DependencySpecifiers.Create(
					productSlug,
					releaseID,
					dependentProductSlug,
					specifier,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete", func() {
		var (
			dependencySpecifierID int
		)

		BeforeEach(func() {
			dependencySpecifierID = 1234
		})

		It("deletes the dependency specifier", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependency_specifiers/%d",
						apiPrefix,
						productSlug,
						releaseID,
						dependencySpecifierID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, nil),
				),
			)

			err := client.DependencySpecifiers.Delete(
				productSlug,
				releaseID,
				dependencySpecifierID,
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
							"%s/products/%s/releases/%d/dependency_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							dependencySpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				err := client.DependencySpecifiers.Delete(
					productSlug,
					releaseID,
					dependencySpecifierID,
				)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf(
							"%s/products/%s/releases/%d/dependency_specifiers/%d",
							apiPrefix,
							productSlug,
							releaseID,
							dependencySpecifierID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.DependencySpecifiers.Delete(
					productSlug,
					releaseID,
					dependencySpecifierID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
