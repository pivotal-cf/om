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
			Expect(err).ToNot(HaveOccurred())
			bar = &fakes.Progress{}
		})

		AfterEach(func() {
			err := os.Remove(outputFile.Name())
			Expect(err).ToNot(HaveOccurred())
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
			Expect(err).ToNot(HaveOccurred())
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
				var roOutputFile *os.File
				BeforeEach(func() {
					var err error
					roOutputFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					err = os.Chmod(roOutputFile.Name(), 0000)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err := os.Remove(roOutputFile.Name())
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
					service := api.NewInstallationService(client, bar)

					err := service.Export(roOutputFile.Name())
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
})
