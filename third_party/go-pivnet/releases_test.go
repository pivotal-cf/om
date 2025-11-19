package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v7/go-pivnetfakes"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - product files", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger
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
		It("returns the releases for the product slug", func() {
			response := `{"releases": [{"id":2,"version":"1.2.3"},{"id": 3, "version": "3.2.1", "_links": {"product_files": {"href":"https://banana.org/cookies/download"}}}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			releases, err := client.Releases.List("banana")
			Expect(err).NotTo(HaveOccurred())
			Expect(releases).To(HaveLen(2))
			Expect(releases[0].ID).To(Equal(2))
			Expect(releases[1].ID).To(Equal(3))
		})

		Context("when specifying a limit", func() {
			It("passes the limit to the API endpoint in the form of query params", func() {
				response := `{"releases": [{"id": 3, "version": "3.2.1", "_links": {"product_files": {"href":"https://banana.org/cookies/download"}}}]}`
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases", "limit=1"),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				releases, err := client.Releases.List("banana", pivnet.QueryParameter{"limit", "1"})
				Expect(err).NotTo(HaveOccurred())
				Expect(releases).To(HaveLen(1))
				Expect(releases[0].ID).To(Equal(3))
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
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Releases.List("banana")
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Releases.List("banana")
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Get", func() {
		It("returns the release for the product slug and releaseID", func() {
			response := `{"id": 3, "version": "3.2.1", "_links": {"product_files": {"href":"https://banana.org/cookies/download"}}}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/3"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			release, err := client.Releases.Get("banana", 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(release.ID).To(Equal(3))
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
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/3"),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Releases.Get("banana", 3)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/3"),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Releases.Get("banana", 3)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Create", func() {
		var (
			releaseVersion      string
			createReleaseConfig pivnet.CreateReleaseConfig
		)

		BeforeEach(func() {
			releaseVersion = "1.2.3.4"

			createReleaseConfig = pivnet.CreateReleaseConfig{
				EULASlug:    "some_eula",
				ReleaseType: "Not a real release",
				Version:     releaseVersion,
				ProductSlug: productSlug,
			}
		})

		Context("when the config is valid", func() {
			type requestBody struct {
				Release      pivnet.Release `json:"release"`
				CopyMetadata bool           `json:"copy_metadata"`
			}

			var (
				expectedReleaseDate string
				expectedRequestBody requestBody

				validResponse string
			)

			BeforeEach(func() {
				expectedReleaseDate = time.Now().Format("2006-01-02")

				expectedRequestBody = requestBody{
					Release: pivnet.Release{
						Availability: "Admins Only",
						OSSCompliant: "confirm",
						ReleaseDate:  expectedReleaseDate,
						ReleaseType:  pivnet.ReleaseType(createReleaseConfig.ReleaseType),
						EULA: &pivnet.EULA{
							Slug: createReleaseConfig.EULASlug,
						},
						Version: createReleaseConfig.Version,
					},
					CopyMetadata: createReleaseConfig.CopyMetadata,
				}

				validResponse = `{"release": {"id": 3, "version": "1.2.3.4"}}`
			})

			It("creates the release with the minimum required fields", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
						ghttp.VerifyJSONRepresenting(&expectedRequestBody),
						ghttp.RespondWith(http.StatusCreated, validResponse),
					),
				)

				release, err := client.Releases.Create(createReleaseConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Version).To(Equal(releaseVersion))
			})

			Context("when the optional release date is present", func() {
				var (
					releaseDate string
				)

				BeforeEach(func() {
					releaseDate = "2015-12-24"

					createReleaseConfig.ReleaseDate = releaseDate
					expectedRequestBody.Release.ReleaseDate = releaseDate
				})

				It("creates the release with the release date field", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
							ghttp.VerifyJSONRepresenting(&expectedRequestBody),
							ghttp.RespondWith(http.StatusCreated, validResponse),
						),
					)

					release, err := client.Releases.Create(createReleaseConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(release.Version).To(Equal(releaseVersion))
				})
			})

			Describe("optional description field", func() {
				var (
					description string
				)

				Context("when the optional description field is present", func() {
					BeforeEach(func() {
						description = "some description"

						createReleaseConfig.Description = description
						expectedRequestBody.Release.Description = description
					})

					It("creates the release with the description field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})

				Context("when the optional description field is not present", func() {
					BeforeEach(func() {
						description = ""

						createReleaseConfig.Description = description
						expectedRequestBody.Release.Description = description
					})

					It("creates the release with an empty description field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})
			})

			Describe("optional release notes URL field", func() {
				var (
					releaseNotesURL string
				)

				Context("when the optional release notes URL field is present", func() {
					BeforeEach(func() {
						releaseNotesURL = "some releaseNotesURL"

						createReleaseConfig.ReleaseNotesURL = releaseNotesURL
						expectedRequestBody.Release.ReleaseNotesURL = releaseNotesURL
					})

					It("creates the release with the release notes URL field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})

				Context("when the optional release notes URL field is not present", func() {
					BeforeEach(func() {
						releaseNotesURL = ""

						createReleaseConfig.ReleaseNotesURL = releaseNotesURL
						expectedRequestBody.Release.ReleaseNotesURL = releaseNotesURL
					})

					It("creates the release with an empty release notes URL field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})
			})

			Describe("optional copy metadata config", func() {
				Context("when the copy metadata config is not present", func() {
					It("creates the release without copying metadata", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})

				Context("when the copy metadata config is true", func() {
					BeforeEach(func() {
						createReleaseConfig.CopyMetadata = true
						expectedRequestBody.CopyMetadata = true
					})

					It("creates the release and copies the metadata", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.Releases.Create(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(releaseVersion))
					})
				})
			})
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
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Releases.Create(createReleaseConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Releases.Create(createReleaseConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update", func() {
		It("submits the updated values for a release with OSS compliance", func() {
			release := pivnet.Release{
				ID:      42,
				Version: "1.2.3.4",
				EULA: &pivnet.EULA{
					Slug: "some-eula",
					ID:   15,
				},
			}

			patchURL := fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, "banana-slug", release.ID)

			response := `{"release": {"id": 42, "version": "1.2.3.4"}}`
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", patchURL),
					ghttp.VerifyJSON(`{"release":{"id": 42, "version": "1.2.3.4", "eula":{"slug":"some-eula","id":15}, "oss_compliant":"confirm"}, "copy_metadata": false}`),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			release, err := client.Releases.Update("banana-slug", release)
			Expect(err).NotTo(HaveOccurred())
			Expect(release.Version).To(Equal("1.2.3.4"))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns the error", func() {
				release := pivnet.Release{ID: 111}
				patchURL := fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, "banana-slug", release.ID)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", patchURL),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.Releases.Update("banana-slug", release)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				release := pivnet.Release{ID: 111}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d",
							apiPrefix,
							"banana-slug",
							release.ID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.Releases.Update("banana-slug", release)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Delete", func() {
		var (
			release pivnet.Release
		)

		BeforeEach(func() {
			release = pivnet.Release{
				ID: 1234,
			}
		})

		It("deletes the release", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, release.ID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)

			err := client.Releases.Delete("banana", release)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, release.ID)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				err := client.Releases.Delete("banana", release)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				release := pivnet.Release{ID: 111}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, release.ID)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.Releases.Delete("banana", release)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
