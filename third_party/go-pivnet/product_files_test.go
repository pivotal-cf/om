package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v7/download"
)

var _ = Describe("PivnetClient - product files", func() {
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

	Describe("List product files", func() {
		var (
			productSlug string

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.ProductFilesResponse{[]pivnet.ProductFile{
				{
					ID:           1234,
					AWSObjectKey: "something",
				},
				{
					ID:           2345,
					AWSObjectKey: "something-else",
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/product_files",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product files without error", func() {
			productFiles, err := client.ProductFiles.List(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFiles).To(HaveLen(2))
			Expect(productFiles[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ProductFiles.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("List product files for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.ProductFilesResponse{[]pivnet.ProductFile{
				{
					ID:           1234,
					AWSObjectKey: "something",
					Links: &pivnet.Links{Download: map[string]string{
						"href": fmt.Sprintf(
							"/products/%s/releases/%d/product_files/%d/download",
							productSlug,
							releaseID,
							1234,
						)},
					},
				},
				{
					ID:           2345,
					AWSObjectKey: "something-else",
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product files without error", func() {
			productFiles, err := client.ProductFiles.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFiles).To(HaveLen(2))
			Expect(productFiles[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ProductFiles.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get Product File", func() {
		var (
			productSlug   string
			productFileID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			productFileID = 1234

			response = pivnet.ProductFileResponse{
				ProductFile: pivnet.ProductFile{
					ID:           productFileID,
					AWSObjectKey: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
							productFileID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			productFile, err := client.ProductFiles.Get(
				productSlug,
				productFileID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(productFileID))
			Expect(productFile.AWSObjectKey).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.Get(
					productSlug,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ProductFiles.Get(
					productSlug,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get product file for release", func() {
		var (
			productSlug   string
			releaseID     int
			productFileID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			productFileID = 1234

			response = pivnet.ProductFileResponse{
				ProductFile: pivnet.ProductFile{
					ID:           productFileID,
					AWSObjectKey: "something",
					Links: &pivnet.Links{Download: map[string]string{
						"href": fmt.Sprintf(
							"/products/%s/releases/%d/product_files/%d/download",
							productSlug,
							releaseID,
							productFileID,
						)},
					},
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files/%d",
							apiPrefix,
							productSlug,
							releaseID,
							productFileID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			productFile, err := client.ProductFiles.GetForRelease(
				productSlug,
				releaseID,
				productFileID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(productFileID))
			Expect(productFile.AWSObjectKey).To(Equal("something"))

			Expect(productFile.Links.Download["href"]).
				To(Equal(fmt.Sprintf(
					"/products/%s/releases/%d/product_files/%d/download",
					productSlug,
					releaseID,
					productFileID,
				)))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.GetForRelease(
					productSlug,
					releaseID,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ProductFiles.GetForRelease(
					productSlug,
					releaseID,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create Product File", func() {
		type requestBody struct {
			ProductFile pivnet.ProductFile `json:"product_file"`
		}

		var (
			createProductFileConfig pivnet.CreateProductFileConfig

			expectedRequestBody requestBody

			productFileResponse pivnet.ProductFileResponse
		)

		BeforeEach(func() {
			createProductFileConfig = pivnet.CreateProductFileConfig{
				ProductSlug:        productSlug,
				AWSObjectKey:       "some-aws-object-key",
				Description:        "some\nmulti-line\ndescription",
				DocsURL:            "some-docs-url",
				FileType:           "some-file-type",
				FileVersion:        "some-file-version",
				IncludedFiles:      []string{"file1", "file2"},
				SHA256:             "some-sha256",
				MD5:                "some-md5",
				Name:               "some-file-name",
				Platforms:          []string{"platform-1", "platform-2"},
				ReleasedAt:         "released-at",
				SystemRequirements: []string{"system-1", "system-2"},
			}

			expectedRequestBody = requestBody{
				ProductFile: pivnet.ProductFile{
					AWSObjectKey:       createProductFileConfig.AWSObjectKey,
					Description:        createProductFileConfig.Description,
					DocsURL:            createProductFileConfig.DocsURL,
					FileType:           createProductFileConfig.FileType,
					FileVersion:        createProductFileConfig.FileVersion,
					IncludedFiles:      createProductFileConfig.IncludedFiles,
					SHA256:             createProductFileConfig.SHA256,
					MD5:                createProductFileConfig.MD5,
					Name:               createProductFileConfig.Name,
					Platforms:          createProductFileConfig.Platforms,
					ReleasedAt:         createProductFileConfig.ReleasedAt,
					SystemRequirements: createProductFileConfig.SystemRequirements,
				},
			}

			productFileResponse = pivnet.ProductFileResponse{
				ProductFile: pivnet.ProductFile{
					ID:                 1234,
					AWSObjectKey:       createProductFileConfig.AWSObjectKey,
					Description:        createProductFileConfig.Description,
					DocsURL:            createProductFileConfig.DocsURL,
					FileType:           createProductFileConfig.FileType,
					FileVersion:        createProductFileConfig.FileVersion,
					IncludedFiles:      createProductFileConfig.IncludedFiles,
					SHA256:             createProductFileConfig.SHA256,
					MD5:                createProductFileConfig.MD5,
					Name:               createProductFileConfig.Name,
					Platforms:          createProductFileConfig.Platforms,
					ReleasedAt:         createProductFileConfig.ReleasedAt,
					SystemRequirements: createProductFileConfig.SystemRequirements,
				}}
		})

		It("creates the product file", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/product_files",
						apiPrefix,
						productSlug,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, productFileResponse),
				),
			)

			productFile, err := client.ProductFiles.Create(createProductFileConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(productFile.ID).To(Equal(1234))
			Expect(productFile).To(Equal(productFileResponse.ProductFile))
		})

		Context("when the server responds with a non-201 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/product_files",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the server responds with a 429 status code", func() {
			It("returns an error indicating the limit was hit", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/product_files",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTooManyRequests, "Retry later"),
					),
				)

				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err.Error()).To(ContainSubstring("You have hit the file creation limit. Please wait before creating more files. Contact pivnet-eng@pivotal.io with additional questions."))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/product_files",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})

		Context("when the aws object key is empty", func() {
			BeforeEach(func() {
				createProductFileConfig = pivnet.CreateProductFileConfig{
					ProductSlug:  productSlug,
					Name:         "some-file-name",
					FileVersion:  "some-file-version",
					AWSObjectKey: "",
				}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("AWS object key"))
			})
		})
	})

	Describe("Update Product File", func() {
		type requestBody struct {
			ProductFile pivnet.ProductFile `json:"product_file"`
		}

		var (
			expectedRequestBody requestBody

			productFile pivnet.ProductFile

			validResponse = `{"product_file":{"id":1234,"docs_url":"http://self-docs.com/","system_requirements": ["1", "2"]}}`
		)

		BeforeEach(func() {
			productFile = pivnet.ProductFile{
				ID:                 1234,
				Description:        "some-description",
				FileVersion:        "some-file-version",
				SHA256:             "some-sha256",
				MD5:                "some-md5",
				Name:               "some-file-name",
				DocsURL:            "http://self-docs.com/",
				SystemRequirements: []string{"1", "2"},
			}

			expectedRequestBody = requestBody{
				ProductFile: pivnet.ProductFile{
					Description:        productFile.Description,
					FileVersion:        productFile.FileVersion,
					SHA256:             productFile.SHA256,
					MD5:                productFile.MD5,
					Name:               productFile.Name,
					DocsURL:            productFile.DocsURL,
					SystemRequirements: productFile.SystemRequirements,
				},
			}
		})

		It("updates the product file with the provided fields", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/products/%s/product_files/%d",
						apiPrefix,
						productSlug,
						productFile.ID,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWith(http.StatusOK, validResponse),
				),
			)

			updatedProductFile, err := client.ProductFiles.Update(productSlug, productFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedProductFile.ID).To(Equal(productFile.ID))
			Expect(updatedProductFile.DocsURL).To(Equal(productFile.DocsURL))
			Expect(updatedProductFile.SystemRequirements).To(ConsistOf("2", "1"))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
							productFile.ID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Update(productSlug, productFile)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
							productFile.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ProductFiles.Update(productSlug, productFile)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete Product File", func() {
		var (
			id = 1234
		)

		It("deletes the product file", func() {
			response := []byte(`{"product_file":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			productFile, err := client.ProductFiles.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(id))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ProductFiles.Delete(productSlug, id)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add Product File to release", func() {
		var (
			productSlug   = "some-product"
			releaseID     = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove Product File from release", func() {
		var (
			productSlug   = "some-product"
			releaseID     = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add Product File to file group", func() {
		var (
			productSlug   = "some-product"
			fileGroupID   = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/add_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/add_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/add_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove Product File from file group", func() {
		var (
			productSlug   = "some-product"
			fileGroupID   = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/remove_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.RemoveFromFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/remove_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.RemoveFromFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/remove_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ProductFiles.RemoveFromFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("ProductFile methods", func() {
		var (
			productFile pivnet.ProductFile
		)

		BeforeEach(func() {
			productFile = pivnet.ProductFile{}
		})

		Describe("DownloadLink", func() {
			var (
				downloadLink string
			)

			BeforeEach(func() {
				downloadLink = "some link"

				productFile.Links = &pivnet.Links{
					Download: map[string]string{
						"href": downloadLink,
					},
				}
			})

			It("returns download link from links map", func() {
				dl, err := productFile.DownloadLink()
				Expect(err).NotTo(HaveOccurred())

				Expect(dl).To(Equal(downloadLink))
			})

			Context("when links are nil", func() {
				BeforeEach(func() {
					productFile.Links = nil
				})

				It("returns error", func() {
					_, err := productFile.DownloadLink()
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(ContainSubstring("empty"))
				})
			})
		})
	})

	Describe("DownloadForRelease", func() {
		var (
			cloudfront    *ghttp.Server
			releaseID     int
			productFileID int

			downloadLink               string
			cloudfrontDownloadLocation string

			downloadLinkResponseBody []byte

			getStatusCode int
			getResponse   interface{}

			downloadLinkResponseStatusCode int
			cloudfrontDownloadPath         string
		)

		BeforeEach(func() {
			releaseID = 1234
			productFileID = 2345

			downloadLink = "/some/download/link"

			downloadLinkResponseBody = []byte("some file contents")

			cloudfront = ghttp.NewServer()

			cloudfrontDownloadLocation = fmt.Sprintf("%s/%s", cloudfront.URL(), "download")

			getStatusCode = http.StatusOK
			getResponse = pivnet.ProductFileResponse{
				pivnet.ProductFile{
					ID:           1234,
					AWSObjectKey: "something",
					Links: &pivnet.Links{
						Download: map[string]string{
							"href": downloadLink,
						},
					},
				},
			}

			downloadLinkResponseStatusCode = http.StatusFound
			cloudfrontDownloadPath = "/download"
		})

		AfterEach(func() {
			cloudfront.Close()
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files/%d",
							apiPrefix,
							productSlug,
							releaseID,
							productFileID,
						),
					),
					ghttp.RespondWithJSONEncoded(getStatusCode, getResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s%s",
						apiPrefix,
						downloadLink,
					)),
					ghttp.RespondWith(downloadLinkResponseStatusCode, []byte(`{}`),
						http.Header{
							"Location": []string{cloudfrontDownloadLocation},
						},
					),
				),
			)

			cloudfront.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("HEAD", "/download"),
					ghttp.RespondWith(http.StatusOK, nil,
						http.Header{
							"Content-Length": []string{"18"},
						},
					),
				),
			)

			cloudfront.RouteToHandler("GET", cloudfrontDownloadPath, ghttp.CombineHandlers(
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ex := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
					matches := ex.FindStringSubmatch(req.Header.Get("Range"))

					start, err := strconv.Atoi(matches[1])
					if err != nil {
						Fail(err.Error())
					}

					end, err := strconv.Atoi(matches[2])
					if err != nil {
						Fail(err.Error())
					}

					w.WriteHeader(http.StatusPartialContent)
					_, err = w.Write(downloadLinkResponseBody[start : end+1])
					Expect(err).NotTo(HaveOccurred())
				}),
			),
			)
		})

		It("writes file contents to provided writer", func() {
			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			tmpLocation, err := download.NewFileInfo(tmpFile)
			Expect(err).NotTo(HaveOccurred())

			err = client.ProductFiles.DownloadForRelease(
				tmpLocation,
				productSlug,
				releaseID,
				productFileID,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(tmpFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(contents).To(Equal(downloadLinkResponseBody))
		})

		Context("when productFile.DownloadLink() returns an error", func() {
			BeforeEach(func() {
				getResponse = pivnet.ProductFileResponse{
					pivnet.ProductFile{
						ID: 1234,
					},
				}
			})

			It("returns the error", func() {
				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				tmpLocation, err := download.NewFileInfo(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				err = client.ProductFiles.DownloadForRelease(
					tmpLocation,
					productSlug,
					releaseID,
					productFileID,
					GinkgoWriter,
				)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when making the request returns an error", func() {
			BeforeEach(func() {
				downloadLinkResponseStatusCode = http.StatusTeapot
			})

			It("forwards the error", func() {
				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				tmpLocation, err := download.NewFileInfo(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				err = client.ProductFiles.DownloadForRelease(
					tmpLocation,
					productSlug,
					releaseID,
					productFileID,
					GinkgoWriter,
				)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the download link returns a forbidden status code", func() {
			BeforeEach(func() {
				cloudfrontDownloadPath = "/valid-download"
			})

			JustBeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s%s",
							apiPrefix,
							downloadLink,
						)),
						ghttp.RespondWith(downloadLinkResponseStatusCode, []byte(`{}`),
							http.Header{
								"Location": []string{fmt.Sprintf("%s/%s", cloudfront.URL(), "valid-download")},
							},
						),
					),
				)

				cloudfront.RouteToHandler("GET", "/download", ghttp.CombineHandlers(
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						ex := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
						matches := ex.FindStringSubmatch(req.Header.Get("Range"))

						start, err := strconv.Atoi(matches[1])
						if err != nil {
							Fail(err.Error())
						}

						end, err := strconv.Atoi(matches[2])
						if err != nil {
							Fail(err.Error())
						}

						if start == 1 && end == 1 {
							w.WriteHeader(http.StatusForbidden)
						} else {
							w.WriteHeader(http.StatusPartialContent)
							_, err = w.Write(downloadLinkResponseBody[start : end+1])
							Expect(err).NotTo(HaveOccurred())
						}
					}),
				))
			})

			It("gets a new cloudfront link from pivnet and retries the download", func() {
				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				tmpLocation, err := download.NewFileInfo(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				err = client.ProductFiles.DownloadForRelease(
					tmpLocation,
					productSlug,
					releaseID,
					productFileID,
					GinkgoWriter,
				)
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())

				Expect(contents).To(Equal(downloadLinkResponseBody))
			})
		})

		Context("when there is an error getting the release", func() {
			BeforeEach(func() {
				getStatusCode = http.StatusTeapot
			})

			It("forwards the error", func() {
				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				tmpLocation, err := download.NewFileInfo(tmpFile)
				Expect(err).NotTo(HaveOccurred())

				err = client.ProductFiles.DownloadForRelease(
					tmpLocation,
					productSlug,
					releaseID,
					productFileID,
					GinkgoWriter,
				)

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
