package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallationsService", func() {
	var (
		client *fakes.HttpClient
		is     api.InstallationsService
	)
	BeforeEach(func() {
		client = &fakes.HttpClient{}
		is = api.NewInstallationsService(client)
	})

	Describe("Trigger", func() {
		It("triggers an installation on an Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"install":{"id":1}}`)),
			}, nil)

			output, err := is.Trigger()

			Expect(err).NotTo(HaveOccurred())
			Expect(output.ID).To(Equal(1))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations"))

			body, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("{}"))
		})

		Context("when an error occurs", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := is.Trigger()
					Expect(err).To(MatchError("could not make api request to installations endpoint: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := is.Trigger()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := is.Trigger()
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("Status", func() {
		It("fetches the status of the installation from the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"status": "running"}`)),
			}, nil)

			output, err := is.Status(3232)

			Expect(err).NotTo(HaveOccurred())
			Expect(output.Status).To(Equal("running"))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations/3232"))
		})

		Context("when an error occurs", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := is.Status(3232)
					Expect(err).To(MatchError("could not make api request to installations status endpoint: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := is.Status(3232)
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := is.Status(3232)
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("Logs", func() {
		It("grabs the logs from the currently running installation", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"logs": "some logs"}`)),
			}, nil)

			output, err := is.Logs(3232)

			Expect(err).NotTo(HaveOccurred())
			Expect(output.Logs).To(Equal("some logs"))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations/3232/logs"))
		})
	})

	Context("when an error occurs", func() {
		Context("when the client has an error during the request", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, errors.New("some error"))

				_, err := is.Logs(3232)
				Expect(err).To(MatchError("could not make api request to installations logs endpoint: some error"))
			})
		})

		Context("when the client returns a non-2XX", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil)

				_, err := is.Logs(3232)
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		Context("when the json cannot be decoded", func() {
			It("returns an error", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader("##################")),
				}, nil)

				_, err := is.Logs(3232)
				Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
			})
		})
	})
})
