package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("DirectorService", func() {
	var (
		client *fakes.HttpClient
		ds     api.DirectorService
	)
	BeforeEach(func() {
		client = &fakes.HttpClient{}
		ds = api.NewDirectorService(client)
	})

	Describe("NetworkAndAZ", func() {
		It("creates an network and az assignment", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := ds.NetworkAndAZ(`{"some-key": "some-value"}`)

			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/network_and_az"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := ds.NetworkAndAZ(`{"some-key": "some-value"}`)

				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := ds.NetworkAndAZ(`{"some-key": "some-value"}`)

				Expect(err).To(MatchError("could not make api request to network and AZ endpoint: api endpoint failed"))
			})
		})
	})

})
