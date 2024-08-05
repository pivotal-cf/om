package api_test

import (
	"net/http"
	"os"
	"strings"

	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallationAssetService", func() {
	var (
		client                 *ghttp.Server
		progressClient         *ghttp.Server
		unauthedProgressClient *ghttp.Server
		service                api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()
		progressClient = ghttp.NewServer()
		unauthedProgressClient = ghttp.NewServer()
		service = api.New(api.ApiInput{
			Client:                 httpClient{serverURI: client.URL()},
			ProgressClient:         httpClient{serverURI: progressClient.URL()},
			UnauthedProgressClient: httpClient{serverURI: unauthedProgressClient.URL()},
		})
	})

	AfterEach(func() {
		client.Close()
		progressClient.Close()
		unauthedProgressClient.Close()
	})

	Describe("DownloadInstallationAssetCollection", func() {
		var (
			outputFile *os.File
		)

		BeforeEach(func() {
			var err error
			outputFile, err = os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(outputFile.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		It("makes a request to export the current Ops Manager installation", func() {
			progressClient.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/installation_asset_collection"),
					ghttp.RespondWith(http.StatusOK, "some-installation"),
				),
			)

			err := service.DownloadInstallationAssetCollection(outputFile.Name())
			Expect(err).ToNot(HaveOccurred())

			By("writing the installation to a local file")
			ins, err := os.ReadFile(outputFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(string(ins)).To(Equal("some-installation"))
		})

		When("the client errors before the request", func() {
			It("returns an error", func() {
				progressClient.Close()

				err := service.DownloadInstallationAssetCollection("fake-file")
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installation_asset_collection endpoint: could not send api request to GET /api/v0/installation_asset_collection")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				progressClient.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installation_asset_collection"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err := service.DownloadInstallationAssetCollection("fake-file")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the output file cannot be written", func() {
			It("returns an error", func() {
				progressClient.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/installation_asset_collection"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.DownloadInstallationAssetCollection("fake-dir/fake-file")
				Expect(err).To(MatchError(ContainSubstring("no such file")))
			})
		})
	})

	Describe("UploadInstallationAssetCollection", func() {
		It("makes a request to import the installation to the Ops Manager", func() {
			unauthedProgressClient.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/installation_asset_collection"),
					ghttp.VerifyBody([]byte("some installation")),
					ghttp.VerifyContentType("some content-type"),
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						Expect(req.ContentLength).To(Equal(int64(17)))

						_, err := w.Write([]byte(`{}`))
						Expect(err).ToNot(HaveOccurred())
					}),
				),
			)

			err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
				ContentLength:   17,
				Installation:    strings.NewReader("some installation"),
				ContentType:     "some content-type",
				PollingInterval: 1,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("an error occurs", func() {
			When("the client errors before the request", func() {
				It("returns an error", func() {
					unauthedProgressClient.Close()

					err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to installation_asset_collection endpoint")))
				})
			})

			When("the api returns a non-200 status code", func() {
				It("returns an error", func() {
					unauthedProgressClient.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v0/installation_asset_collection"),
							ghttp.RespondWith(http.StatusTeapot, `{}`),
						),
					)

					err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("DeleteInstallationAssetCollection", func() {
		It("makes a request to delete the installation on the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/installation_asset_collection"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{"errands": {}}`),
					ghttp.RespondWith(http.StatusOK, `{
						"install": {
							"id": 12
						}
					}`),
				),
			)

			output, err := service.DeleteInstallationAssetCollection()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ID).To(Equal(12))
		})

		It("gracefully quits when there is no installation to delete", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/installation_asset_collection"),
					ghttp.RespondWith(http.StatusGone, `{}`),
				),
			)

			output, err := service.DeleteInstallationAssetCollection()
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{}))
		})

		When("the client errors before the request", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.DeleteInstallationAssetCollection()
				Expect(err).To(MatchError(ContainSubstring("could not make api request to installation_asset_collection endpoint: could not send api request to DELETE /api/v0/installation_asset_collection")))
			})
		})

		When("the api returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/installation_asset_collection"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.DeleteInstallationAssetCollection()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the api response cannot be unmarshaled", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/installation_asset_collection"),
						ghttp.RespondWith(http.StatusOK, `%%%`),
					),
				)

				_, err := service.DeleteInstallationAssetCollection()
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
})
