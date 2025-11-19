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

var _ = Describe("PivnetClient - FileGroup", func() {
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
		It("returns all FileGroups", func() {
			response := pivnet.FileGroupsResponse{
				[]pivnet.FileGroup{
					{
						ID:   1234,
						Name: "Some file group",
					},
					{
						ID:   2345,
						Name: "Some other file group",
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/file_groups", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			fileGroups, err := client.FileGroups.List(productSlug)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroups).To(HaveLen(2))

			Expect(fileGroups[0].ID).To(Equal(fileGroups[0].ID))
			Expect(fileGroups[0].Name).To(Equal(fileGroups[0].Name))
			Expect(fileGroups[1].ID).To(Equal(fileGroups[1].ID))
			Expect(fileGroups[1].Name).To(Equal(fileGroups[1].Name))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/file_groups", apiPrefix, productSlug)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.FileGroups.List(productSlug)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/file_groups", apiPrefix, productSlug)),
						ghttp.RespondWith(http.StatusOK, "%%%"),
					),
				)

				_, err := client.FileGroups.List(productSlug)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("List for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.FileGroupsResponse{[]pivnet.FileGroup{
				{
					ID:   1234,
					Name: "something",
				},
				{
					ID:   2345,
					Name: "something-else",
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
							"%s/products/%s/releases/%d/file_groups",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			fileGroups, err := client.FileGroups.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroups).To(HaveLen(2))
			Expect(fileGroups[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.FileGroups.ListForRelease(
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
				_, err := client.FileGroups.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get File group", func() {
		var (
			productSlug string
			fileGroupID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			fileGroupID = 1234

			response = pivnet.FileGroup{
				ID:   fileGroupID,
				Name: "something",
			}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the file group without error", func() {
			fileGroup, err := client.FileGroups.Get(
				productSlug,
				fileGroupID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroup.ID).To(Equal(fileGroupID))
			Expect(fileGroup.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.FileGroups.Get(
					productSlug,
					fileGroupID,
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
				_, err := client.FileGroups.Get(
					productSlug,
					fileGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create", func() {
		var (
			name string

			expectedRequestBody string

			returnedFileGroup pivnet.FileGroup
		)

		BeforeEach(func() {
			name = "some name"

			expectedRequestBody = fmt.Sprintf(
				`{"file_group":{"name":"%s"}}`,
				name,
			)
		})

		JustBeforeEach(func() {
			returnedFileGroup = pivnet.FileGroup{
				ID:   1234,
				Name: name,
			}
		})

		It("creates new file group without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/file_groups",
						apiPrefix,
						productSlug,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, returnedFileGroup),
				),
			)

			config := pivnet.CreateFileGroupConfig{productSlug, name}
			fileGroup, err := client.FileGroups.Create(config)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroup.ID).To(Equal(returnedFileGroup.ID))
			Expect(fileGroup.Name).To(Equal(name))
		})

		Context("when the server responds with a non-201 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/file_groups",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				config := pivnet.CreateFileGroupConfig{productSlug, name}
				_, err := client.FileGroups.Create(config)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/file_groups",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				config := pivnet.CreateFileGroupConfig{productSlug, name}
				_, err := client.FileGroups.Create(config)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update", func() {
		var (
			fileGroup pivnet.FileGroup

			expectedRequestBody string

			response pivnet.FileGroup
		)

		BeforeEach(func() {
			fileGroup = pivnet.FileGroup{
				ID:   1234,
				Name: "some name",
			}

			expectedRequestBody = fmt.Sprintf(
				`{"file_group":{"name":"%s"}}`,
				fileGroup.Name,
			)

			response = fileGroup
		})

		It("returns without error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/products/%s/file_groups/%d",
						apiPrefix,
						productSlug,
						fileGroup.ID,
					)),
					ghttp.VerifyJSON(expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			returned, err := client.FileGroups.Update(productSlug, fileGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned.ID).To(Equal(fileGroup.ID))
			Expect(returned.Name).To(Equal(fileGroup.Name))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.FileGroups.Update(productSlug, fileGroup)

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroup.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.FileGroups.Update(productSlug, fileGroup)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete File Group", func() {
		var (
			id = 1234
		)

		It("deletes the file group", func() {
			response := []byte(`{"id":1234}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/file_groups/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			fileGroup, err := client.FileGroups.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroup.ID).To(Equal(id))
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
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/file_groups/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.FileGroups.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/file_groups/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.FileGroups.Delete(productSlug, id)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add File Group", func() {
		var (
			productSlug = "some-product"
			releaseID   = 2345
			fileGroupID = 3456

			expectedRequestBody = `{"file_group":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.FileGroups.AddToRelease(
					productSlug,
					releaseID,
					fileGroupID,
				)
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
							"%s/products/%s/releases/%d/add_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.FileGroups.AddToRelease(productSlug, releaseID, fileGroupID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.FileGroups.AddToRelease(productSlug, releaseID, fileGroupID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove File Group", func() {
		var (
			productSlug = "some-product"
			releaseID   = 2345
			fileGroupID = 3456

			expectedRequestBody = `{"file_group":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.FileGroups.RemoveFromRelease(
					productSlug,
					releaseID,
					fileGroupID,
				)
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
							"%s/products/%s/releases/%d/remove_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.FileGroups.RemoveFromRelease(productSlug, releaseID, fileGroupID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_file_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.FileGroups.RemoveFromRelease(productSlug, releaseID, fileGroupID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
