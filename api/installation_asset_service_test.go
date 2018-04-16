package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallationAssetService", func() {
	Describe("DownloadInstallationAssetCollection", func() {
		var (
			progressClient *fakes.HttpClient
			outputFile     *os.File
			bar            *fakes.Progress
			liveWriter     *fakes.LiveWriter
		)

		BeforeEach(func() {
			var err error
			progressClient = &fakes.HttpClient{}
			liveWriter = &fakes.LiveWriter{}
			outputFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			bar = &fakes.Progress{}
		})

		AfterEach(func() {
			err := os.Remove(outputFile.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("makes a request to export the current Ops Manager installation", func() {
			progressClient.DoReturns(&http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: int64(len([]byte("some-installation"))),
				Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
			}, nil)

			service := api.NewInstallationAssetService(nil, progressClient)

			err := service.DownloadInstallationAssetCollection(outputFile.Name(), 1)
			Expect(err).NotTo(HaveOccurred())

			By("posting to the installation_asset_collection endpoint")
			request := progressClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/installation_asset_collection"))

			By("writing the installation to a local file")
			ins, err := ioutil.ReadFile(outputFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(ins)).To(Equal("some-installation"))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.DownloadInstallationAssetCollection("fake-file", 1)
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.DownloadInstallationAssetCollection("fake-file", 1)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the output file cannot be written", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.DownloadInstallationAssetCollection("fake-dir/fake-file", 1)
					Expect(err).To(MatchError(ContainSubstring("no such file")))
				})
			})

			Context("when the response length doesn't match the number of bytes copied", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(strings.NewReader("{}")),
						ContentLength: 50,
					}, nil)
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.DownloadInstallationAssetCollection(outputFile.Name(), 1)
					Expect(err).To(MatchError(ContainSubstring("invalid response length")))
				})
			})
		})
	})

	Describe("UploadInstallationAssetCollection", func() {
		var (
			progressClient *fakes.HttpClient
		)

		BeforeEach(func() {
			progressClient = &fakes.HttpClient{}
		})

		It("makes a request to import the installation to the Ops Manager", func() {
			progressClient.DoStub = func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader("{}")),
				}, nil
			}

			service := api.NewInstallationAssetService(nil, progressClient)

			err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
				ContentLength:   10,
				Installation:    strings.NewReader("some installation"),
				ContentType:     "some content-type",
				PollingInterval: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			request := progressClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/installation_asset_collection"))
			Expect(request.ContentLength).To(Equal(int64(10)))
			Expect(request.Header.Get("Content-Type")).To(Equal("some content-type"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal("some installation"))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationAssetService(nil, progressClient)

					err := service.UploadInstallationAssetCollection(api.ImportInstallationInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("DeleteInstallationAssetCollection", func() {
		var (
			client *fakes.HttpClient
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
		})

		It("makes a request to delete the installation on the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				time.Sleep(1 * time.Second)
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"install": {
							"id": 12
						}
					}`)),
				}, nil
			}

			service := api.NewInstallationAssetService(client, nil)

			output, err := service.DeleteInstallationAssetCollection()
			Expect(err).NotTo(HaveOccurred())
			Expect(output.ID).To(Equal(12))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("DELETE"))
			Expect(request.URL.Path).To(Equal("/api/v0/installation_asset_collection"))
			Expect(request.Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal(`{"errands": {}}`))
		})

		It("gracefully quits when there is no installation to delete", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				time.Sleep(1 * time.Second)
				return &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					StatusCode: http.StatusGone,
				}, nil
			}

			service := api.NewInstallationAssetService(client, nil)

			output, err := service.DeleteInstallationAssetCollection()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{}))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(client, nil)

					_, err := service.DeleteInstallationAssetCollection()
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationAssetService(client, nil)

					_, err := service.DeleteInstallationAssetCollection()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the api response cannot be unmarshaled", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("%%%")),
					}, nil)
					service := api.NewInstallationAssetService(client, nil)

					_, err := service.DeleteInstallationAssetCollection()
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
