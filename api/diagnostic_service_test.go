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

var _ = Describe("DiagnosticService", func() {
	var client *fakes.HttpClient

	BeforeEach(func() {
		client = &fakes.HttpClient{}
	})

	Describe("Report", func() {
		It("returns a diagnostic report", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{"stemcells":["some-stemcell", "some-other-stemcell"]}`)),
			}, nil)

			service := api.NewDiagnosticService(client)
			report, err := service.Report()
			Expect(err).NotTo(HaveOccurred())

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("GET"))
			Expect(request.URL.Path).To(Equal("/api/v0/diagnostic_report"))

			Expect(report.Stemcells).To(Equal([]string{"some-stemcell", "some-other-stemcell"}))
		})

		Context("when an error occurs", func() {
			Context("when the client fails before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))

					service := api.NewDiagnosticService(client)
					_, err := service.Report()
					Expect(err).To(MatchError("could not make api request to diagnostic_report endpoint: some error"))
				})
			})

			Context("when the server returns a non-2XX status", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					service := api.NewDiagnosticService(client)
					_, err := service.Report()
					Expect(err).NotTo(MatchError("request failed: unexpected response"))
				})
			})

			Context("when invalid json is returned", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`$$$$$`)),
					}, nil)

					service := api.NewDiagnosticService(client)
					_, err := service.Report()
					Expect(err).NotTo(MatchError("invalid json received from server"))
				})
			})
		})
	})
})
