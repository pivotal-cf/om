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

var _ = Describe("PivnetClient - product", func() {
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

	Describe("Get", func() {
		var (
			slug         = "my-product"
			s3Directory  = "my-product/path"
			isPksProduct = true
		)

		Context("when the product can be found", func() {
			It("returns the located product", func() {
				response := fmt.Sprintf(`{"id": 3, "slug": "%s"}`, slug)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				product, err := client.Products.Get(slug)
				Expect(err).NotTo(HaveOccurred())
				Expect(product.Slug).To(Equal(slug))
				Expect(product.S3Directory).To(BeNil())
			})

			Context("when the product includes the s3Directory", func() {
				It("contains the s3 prefix path", func() {
					response := fmt.Sprintf(`{"id": 3, "slug": "%s", "s3_directory": { "path": "%s" }}`, slug, s3Directory)

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf(
								"%s/products/%s",
								apiPrefix,
								slug)),
							ghttp.RespondWith(http.StatusOK, response),
						),
					)

					product, err := client.Products.Get(slug)
					Expect(err).NotTo(HaveOccurred())
					Expect(product.S3Directory).NotTo(BeNil())
					Expect(product.S3Directory.Path).To(Equal(s3Directory))
				})
			})

			Context("when the product is installable on PKS", func() {
				It("contains installs_on_pks field", func() {
					response := fmt.Sprintf(`{"id": 3, "slug": "%s", "s3_directory": { "path": "%s" }, "installs_on_pks": %t}`, slug, s3Directory, isPksProduct)

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf(
								"%s/products/%s",
								apiPrefix,
								slug)),
							ghttp.RespondWith(http.StatusOK, response),
						),
					)

					product, err := client.Products.Get(slug)
					Expect(err).NotTo(HaveOccurred())
					Expect(product.S3Directory).NotTo(BeNil())
					Expect(product.S3Directory.Path).To(Equal(s3Directory))
					Expect(product.InstallsOnPks).To(Equal(isPksProduct))
				})
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
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Products.Get(slug)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s",
							apiPrefix,
							slug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Products.Get(slug)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("List", func() {
		var (
			slug = "my-product"
		)

		Context("when the products can be found", func() {
			It("returns the products", func() {
				response := fmt.Sprintf(`{"products":[{"id": 3, "slug": "%s"},{"id": 4, "slug": "bar"}]}`, slug)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products",
							apiPrefix)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				products, err := client.Products.List()
				Expect(err).NotTo(HaveOccurred())

				Expect(products).To(HaveLen(2))
				Expect(products[0].Slug).To(Equal(slug))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products",
							apiPrefix,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Products.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products",
							apiPrefix,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Products.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("SlugAlias", func() {
		Context("when product can be found", func() {
			var (
				productSlug = "product-slug"
				productSlugAlias = "product-slug-alias"
			)

			It("returns list of all the slugs associated to the product", func() {
				response := fmt.Sprintf(`{"slugs": ["%[1]v", "%[2]v"], "current_slug": "%[1]v"}`, productSlug, productSlugAlias)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/slug_alias",
							apiPrefix,
							productSlug)),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				slugAliasResponse, err := client.Products.SlugAlias(productSlug)
				Expect(err).NotTo(HaveOccurred())
				Expect(slugAliasResponse.Slugs).To(HaveLen(2))
				Expect(slugAliasResponse.Slugs).To(ConsistOf(productSlugAlias, productSlug))
				Expect(slugAliasResponse.CurrentSlug).To(Equal(productSlug))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/slug_alias",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Products.SlugAlias(productSlug)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})
})
