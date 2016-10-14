package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallationService", func() {
	Describe("Export", func() {
		var (
			client     *fakes.HttpClient
			outputFile *os.File
			bar        *fakes.Progress
		)

		BeforeEach(func() {
			var err error
			client = &fakes.HttpClient{}
			outputFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			bar = &fakes.Progress{}
		})

		AfterEach(func() {
			err := os.Remove(outputFile.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("makes a request to export the current OpsManager installation", func() {
			client.DoReturns(&http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: 22,
				Body:          ioutil.NopCloser(strings.NewReader("some-installation")),
			}, nil)

			bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
			service := api.NewInstallationService(client, bar)

			err := service.Export(outputFile.Name())
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

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationService(client, bar)

					err := service.Export("fake-file")
					Expect(err).To(MatchError("could not make api request to installation_asset_collection endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewInstallationService(client, bar)

					err := service.Export("fake-file")
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
					service := api.NewInstallationService(client, bar)

					err := service.Export("fake-dir/fake-file")
					Expect(err).To(MatchError(ContainSubstring("no such file")))
				})
			})
		})
	})

	Describe("Import", func() {
		var (
			client *fakes.HttpClient
			bar    *fakes.Progress
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			bar = &fakes.Progress{}
		})

		It("makes a request to import the installation to the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil)

			bar.NewBarReaderReturns(strings.NewReader("some other installation"))
			service := api.NewInstallationService(client, bar)

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
			Expect(bar.EndCallCount()).To(Equal(1))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewInstallationService(client, bar)

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
					service := api.NewInstallationService(client, bar)

					err := service.Import(api.ImportInstallationInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})
})
