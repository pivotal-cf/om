package api_test

import (
	"errors"
	"fmt"
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
	Describe("Export", func() {
		var (
			client     *fakes.HttpClient
			outputFile *os.File
			bar        *fakes.Progress
			liveWriter *fakes.LiveWriter
		)

		BeforeEach(func() {
			var err error
			client = &fakes.HttpClient{}
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
			client.DoReturns(&http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: 22,
				Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
			}, nil)

			bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
			service := api.NewInstallationAssetService(client, bar, liveWriter)

			err := service.Export(outputFile.Name(), 1)
			Expect(err).NotTo(HaveOccurred())

			By("posting to the installation_asset_collection endpoint")
			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/installation_asset_collection"))

			By("writing the installation to a local file")
			ins, err := ioutil.ReadFile(outputFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(string(ins)).To(Equal("some-fake-installation"))

			newReaderContent, err := ioutil.ReadAll(bar.NewBarReaderArgsForCall(0))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(newReaderContent)).To(Equal("some-installation"))
			Expect(bar.SetTotalArgsForCall(0)).To(BeNumerically("==", 22))
			Expect(bar.KickoffCallCount()).To(Equal(1))
			Expect(bar.EndCallCount()).To(Equal(1))
		})

		It("logs while waiting for a response from the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/installation_asset_collection" {
					time.Sleep(5 * time.Second)
					return &http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
						ContentLength: 22,
					}, nil
				}
				return nil, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
			service := api.NewInstallationAssetService(client, bar, liveWriter)

			err := service.Export(outputFile.Name(), 1)
			Expect(err).NotTo(HaveOccurred())

			By("starting the live log writer")
			Expect(liveWriter.StartCallCount()).To(Equal(1))

			By("writing to the live log writer")
			Expect(liveWriter.WriteCallCount()).To(Equal(5))
			for i := 0; i < 5; i++ {
				Expect(string(liveWriter.WriteArgsForCall(i))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", i+1)))
			}

			By("flushing the live log writer")
			Expect(liveWriter.StopCallCount()).To(Equal(1))
		})

		Context("when the polling interval is specified", func() {
			It("prints logs at the interval", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.URL.Path == "/api/v0/installation_asset_collection" {
						time.Sleep(6 * time.Second)
						return &http.Response{
							StatusCode:    http.StatusOK,
							Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
							ContentLength: 22,
						}, nil
					}
					return nil, nil
				}

				bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
				service := api.NewInstallationAssetService(client, bar, liveWriter)

				err := service.Export(outputFile.Name(), 2)
				Expect(err).NotTo(HaveOccurred())

				By("starting the live log writer")
				Expect(liveWriter.StartCallCount()).To(Equal(1))

				By("writing to the live log writer")
				Expect(liveWriter.WriteCallCount()).To(Equal(3))
				for argCall, time := 0, 2; argCall < 3; argCall, time = argCall+1, time+2 {
					Expect(string(liveWriter.WriteArgsForCall(argCall))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", time)))
				}

				By("flushing the live log writer")
				Expect(liveWriter.StopCallCount()).To(Equal(1))

			})
			Context("when the polling interval is higher than the time it takes to export the installation", func() {
				It("exits when the export finishes", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						if req.URL.Path == "/api/v0/installation_asset_collection" {
							time.Sleep(2 * time.Second)
							return &http.Response{
								StatusCode:    http.StatusOK,
								Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
								ContentLength: 22,
							}, nil
						}
						return nil, nil
					}

					bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					By("exiting when the export is finished")
					var (
						done bool
						err  error
					)
					go func() {
						err = service.Export(outputFile.Name(), 20)
						done = true
					}()
					Eventually(func() bool {
						return done
					}, 3*time.Second).Should(BeTrue())
					Expect(err).ToNot(HaveOccurred())

					By("starting the live log writer")
					Expect(liveWriter.StartCallCount()).To(Equal(1))

					By("not writing to the live log writer")
					Expect(liveWriter.WriteCallCount()).To(Equal(0))

					By("flushing the live log writer")
					Expect(liveWriter.StopCallCount()).To(Equal(1))

				})
			})
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Export("fake-file", 1)
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Export("fake-file", 1)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the output file cannot be written", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Export("fake-dir/fake-file", 1)
					Expect(err).To(MatchError(ContainSubstring("no such file")))
				})
			})

			Context("when the response length doesn't match the number of bytes copied", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(strings.NewReader("{}")),
						ContentLength: 50,
					}, nil)
					bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Export(outputFile.Name(), 1)
					Expect(err).To(MatchError(ContainSubstring("invalid response length")))
				})
			})
		})
	})

	Describe("Import", func() {
		var (
			client     *fakes.HttpClient
			bar        *fakes.Progress
			liveWriter *fakes.LiveWriter
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			liveWriter = &fakes.LiveWriter{}
			bar = &fakes.Progress{}
		})

		It("makes a request to import the installation to the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader("{}")),
				}, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some other installation"))
			service := api.NewInstallationAssetService(client, bar, liveWriter)

			err := service.Import(api.ImportInstallationInput{
				ContentLength: 10,
				Installation:  strings.NewReader("some installation"),
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/installation_asset_collection"))
			Expect(request.ContentLength).To(Equal(int64(10)))
			Expect(request.Header.Get("Content-Type")).To(Equal("some content-type"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal("some other installation"))

			newReaderContent, err := ioutil.ReadAll(bar.NewBarReaderArgsForCall(0))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(newReaderContent)).To(Equal("some installation"))
			Expect(bar.SetTotalArgsForCall(0)).To(BeNumerically("==", 10))
			Expect(bar.KickoffCallCount()).To(Equal(1))
			By("ending the progress bar")
			Expect(bar.EndCallCount()).To(Equal(1))
		})

		It("logs while waiting for a response from the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/installation_asset_collection" {
					time.Sleep(5 * time.Second)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil
				}
				return nil, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
			service := api.NewInstallationAssetService(client, bar, liveWriter)

			err := service.Import(api.ImportInstallationInput{
				ContentLength: 10,
				Installation:  strings.NewReader("some installation"),
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())

			By("starting the live log writer")
			Expect(liveWriter.StartCallCount()).To(Equal(1))

			By("writing to the live log writer")
			Expect(liveWriter.WriteCallCount()).To(Equal(5))
			for i := 0; i < 5; i++ {
				Expect(string(liveWriter.WriteArgsForCall(i))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", i+1)))
			}

			By("flushing the live log writer")
			Expect(liveWriter.StopCallCount()).To(Equal(1))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Import(api.ImportInstallationInput{})
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationAssetService(client, bar, liveWriter)

					err := service.Import(api.ImportInstallationInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("Delete", func() {
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

			service := api.NewInstallationAssetService(client, nil, nil)

			output, err := service.Delete()
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

			service := api.NewInstallationAssetService(client, nil, nil)

			output, err := service.Delete()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{}))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationAssetService(client, nil, nil)

					_, err := service.Delete()
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationAssetService(client, nil, nil)

					_, err := service.Delete()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the api response cannot be unmarshaled", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("%%%")),
					}, nil)
					service := api.NewInstallationAssetService(client, nil, nil)

					_, err := service.Delete()
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
