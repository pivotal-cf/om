package pivnet_test

import (
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - artifact references", func() {
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

	Describe("List artifact references", func() {
		var (
			productSlug string

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.ArtifactReferencesResponse{[]pivnet.ArtifactReference{
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
							"%s/products/%s/artifact_references",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the artifact references without error", func() {
			artifactReferences, err := client.ArtifactReferences.List(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReferences).To(HaveLen(2))
			Expect(artifactReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ArtifactReferences.List(
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
				_, err := client.ArtifactReferences.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("List artifact references for specific digest", func() {
		var (
			productSlug string
			digest      string

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			digest = "sha256:digest"

			response = pivnet.ArtifactReferencesResponse{[]pivnet.ArtifactReference{
				{
					ID:   1234,
					Name: "something",
				},
				{
					ID:   2345,
					Name: "something-else",
					ReleaseVersions: []string{"1.0.0","1.2.3"},
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
							"%s/products/%s/artifact_references",
							apiPrefix,
							productSlug,
						),
						"digest=sha256:digest",
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the artifact references without error", func() {
			artifactReferences, err := client.ArtifactReferences.ListForDigest(
				productSlug,
				digest,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReferences).To(HaveLen(2))
			Expect(artifactReferences[0].ID).To(Equal(1234))
			Expect(artifactReferences[1].ReleaseVersions).To(HaveLen(2))
			Expect(artifactReferences[1].ReleaseVersions[0]).To(Equal("1.0.0"))
			Expect(artifactReferences[1].ReleaseVersions[1]).To(Equal("1.2.3"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ArtifactReferences.ListForDigest(
					productSlug,
					digest,
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
				_, err := client.ArtifactReferences.ListForDigest(
					productSlug,
					digest,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("List artifact references for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.ArtifactReferencesResponse{[]pivnet.ArtifactReference{
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
							"%s/products/%s/releases/%d/artifact_references",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the artifact references without error", func() {
			artifactReferences, err := client.ArtifactReferences.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReferences).To(HaveLen(2))
			Expect(artifactReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ArtifactReferences.ListForRelease(
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
				_, err := client.ArtifactReferences.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get artifact Reference", func() {
		var (
			productSlug         string
			artifactReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			artifactReferenceID = 1234

			response = pivnet.ArtifactReferenceResponse{
				ArtifactReference: pivnet.ArtifactReference{
					ID:   artifactReferenceID,
					Name: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/artifact_references/%d",
							apiPrefix,
							productSlug,
							artifactReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the artifact reference without error", func() {
			artifactReference, err := client.ArtifactReferences.Get(
				productSlug,
				artifactReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReference.ID).To(Equal(artifactReferenceID))
			Expect(artifactReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ArtifactReferences.Get(
					productSlug,
					artifactReferenceID,
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
				_, err := client.ArtifactReferences.Get(
					productSlug,
					artifactReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get artifact reference for release", func() {
		var (
			productSlug         string
			releaseID           int
			artifactReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			artifactReferenceID = 1234

			response = pivnet.ArtifactReferenceResponse{
				ArtifactReference: pivnet.ArtifactReference{
					ID:   artifactReferenceID,
					Name: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/artifact_references/%d",
							apiPrefix,
							productSlug,
							releaseID,
							artifactReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the artifact reference without error", func() {
			artifactReference, err := client.ArtifactReferences.GetForRelease(
				productSlug,
				releaseID,
				artifactReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReference.ID).To(Equal(artifactReferenceID))
			Expect(artifactReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ArtifactReferences.GetForRelease(
					productSlug,
					releaseID,
					artifactReferenceID,
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
				_, err := client.ArtifactReferences.GetForRelease(
					productSlug,
					releaseID,
					artifactReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create Artifact Reference", func() {
		type requestBody struct {
			ArtifactReference pivnet.ArtifactReference `json:"artifact_reference"`
		}

		var (
			createArtifactReferenceConfig pivnet.CreateArtifactReferenceConfig

			expectedRequestBody requestBody

			artifactReferenceResponse pivnet.ArtifactReferenceResponse
		)

		BeforeEach(func() {
			createArtifactReferenceConfig = pivnet.CreateArtifactReferenceConfig{
				ProductSlug:        productSlug,
				Description:        "some\nmulti-line\ndescription",
				Digest:             "sha256:mydigest",
				DocsURL:            "some-docs-url",
				ArtifactPath:       "my/path:123",
				Name:               "some-artifact-name",
				SystemRequirements: []string{"system-1", "system-2"},
			}

			expectedRequestBody = requestBody{
				ArtifactReference: pivnet.ArtifactReference{
					Description:        createArtifactReferenceConfig.Description,
					Digest:             createArtifactReferenceConfig.Digest,
					DocsURL:            createArtifactReferenceConfig.DocsURL,
					ArtifactPath:       createArtifactReferenceConfig.ArtifactPath,
					Name:               createArtifactReferenceConfig.Name,
					SystemRequirements: createArtifactReferenceConfig.SystemRequirements,
				},
			}

			artifactReferenceResponse = pivnet.ArtifactReferenceResponse{
				ArtifactReference: pivnet.ArtifactReference{
					ID:                 1234,
					Description:        createArtifactReferenceConfig.Description,
					Digest:             createArtifactReferenceConfig.Digest,
					DocsURL:            createArtifactReferenceConfig.DocsURL,
					ArtifactPath:       createArtifactReferenceConfig.ArtifactPath,
					Name:               createArtifactReferenceConfig.Name,
					SystemRequirements: createArtifactReferenceConfig.SystemRequirements,
					ReplicationStatus:  pivnet.InProgress,
				}}
		})

		It("creates the artifact reference", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/artifact_references",
						apiPrefix,
						productSlug,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, artifactReferenceResponse),
				),
			)

			artifactReference, err := client.ArtifactReferences.Create(createArtifactReferenceConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(artifactReference.ID).To(Equal(1234))
			Expect(artifactReference).To(Equal(artifactReferenceResponse.ArtifactReference))
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
							"%s/products/%s/artifact_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ArtifactReferences.Create(createArtifactReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the server responds with a 429 status code", func() {
			It("returns an error indicating the limit was hit", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/artifact_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTooManyRequests, "Retry later"),
					),
				)

				_, err := client.ArtifactReferences.Create(createArtifactReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("You have hit the artifact reference creation limit. Please wait before creating more artifact references. Contact pivnet-eng@pivotal.io with additional questions."))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/artifact_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ArtifactReferences.Create(createArtifactReferenceConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update artifact reference", func() {
		type requestBody struct {
			ArtifactReference pivnet.ArtifactReference `json:"artifact_reference"`
		}

		var (
			expectedRequestBody        requestBody
			artifactReference          pivnet.ArtifactReference
			updateArtifactReferenceUrl string
			validResponse              = `{"artifact_reference":{"id":1234, "docs_url":"example.io", "system_requirements": ["1", "2"], "replication_status": "in_progress"}}`
		)

		BeforeEach(func() {
			artifactReference = pivnet.ArtifactReference{
				ID:                 1234,
				ArtifactPath:       "some/path",
				Description:        "Avast! Pieces o' passion are forever fine.",
				Digest:             "some-sha265",
				DocsURL:            "example.io",
				Name:               "turpis-hercle",
				SystemRequirements: []string{"1", "2"},
			}

			expectedRequestBody = requestBody{
				ArtifactReference: pivnet.ArtifactReference{
					Description:        artifactReference.Description,
					Name:               artifactReference.Name,
					DocsURL:            artifactReference.DocsURL,
					SystemRequirements: artifactReference.SystemRequirements,
				},
			}

			updateArtifactReferenceUrl = fmt.Sprintf(
				"%s/products/%s/artifact_references/%d",
				apiPrefix,
				productSlug,
				artifactReference.ID,
			)

		})

		It("updates the artifact reference with the provided fields", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateArtifactReferenceUrl),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWith(http.StatusOK, validResponse),
				),
			)

			updatedArtifactReference, err := client.ArtifactReferences.Update(productSlug, artifactReference)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedArtifactReference.ID).To(Equal(artifactReference.ID))
			Expect(updatedArtifactReference.DocsURL).To(Equal(artifactReference.DocsURL))
			Expect(updatedArtifactReference.SystemRequirements).To(ConsistOf("2", "1"))
			Expect(updatedArtifactReference.ReplicationStatus).To(Equal(pivnet.InProgress))
		})

		It("forwards the server-side error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateArtifactReferenceUrl),
					ghttp.RespondWithJSONEncoded(http.StatusTeapot,
						pivnetErr{Message: "Meet, scotty, powerdrain!"}),
				),
			)

			_, err := client.ArtifactReferences.Update(productSlug, artifactReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("scotty"))
		})

		It("forwards the unmarshalling error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateArtifactReferenceUrl),
					ghttp.RespondWith(http.StatusTeapot, "<NOT></JSON>"),
				),
			)
			_, err := client.ArtifactReferences.Update(productSlug, artifactReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})

	Describe("Delete Artifact Reference", func() {
		var (
			id = 1234
		)

		It("deletes the artifact reference", func() {
			response := []byte(`{"artifact_reference":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/artifact_references/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			artifactReference, err := client.ArtifactReferences.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(artifactReference.ID).To(Equal(id))
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
							fmt.Sprintf("%s/products/%s/artifact_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ArtifactReferences.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/artifact_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ArtifactReferences.Delete(productSlug, id)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add Artifact Reference to release", func() {
		var (
			productSlug         = "some-product"
			releaseID           = 2345
			artifactReferenceID = 3456

			expectedRequestBody = `{"artifact_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ArtifactReferences.AddToRelease(productSlug, releaseID, artifactReferenceID)
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
							"%s/products/%s/releases/%d/add_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ArtifactReferences.AddToRelease(productSlug, releaseID, artifactReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ArtifactReferences.AddToRelease(productSlug, releaseID, artifactReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove Artifact Reference from release", func() {
		var (
			productSlug         = "some-product"
			releaseID           = 2345
			artifactReferenceID = 3456

			expectedRequestBody = `{"artifact_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ArtifactReferences.RemoveFromRelease(productSlug, releaseID, artifactReferenceID)
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
							"%s/products/%s/releases/%d/remove_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ArtifactReferences.RemoveFromRelease(productSlug, releaseID, artifactReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_artifact_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ArtifactReferences.RemoveFromRelease(productSlug, releaseID, artifactReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
