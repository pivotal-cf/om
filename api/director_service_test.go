package api_test

import (
	"encoding/json"
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
		client          *fakes.HttpClient
		directorService api.DirectorService
	)
	BeforeEach(func() {
		client = &fakes.HttpClient{}
		directorService = api.NewDirectorService(client)
	})

	Describe("NetworkAndAZ", func() {
		It("creates an network and az assignment", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := directorService.NetworkAndAZ(api.NetworkAndAZConfiguration{
				NetworkAZ: json.RawMessage(`{
					"network": {"name": "network_name"},
					"singleton_availability_zone": {"name": "availability_zone_name"}
				}`),
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/network_and_az"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"network_and_az": {
				   "network": {
					 "name": "network_name"
				   },
				   "singleton_availability_zone": {
					 "name": "availability_zone_name"
				   }
				}
			}`))
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := directorService.NetworkAndAZ(api.NetworkAndAZConfiguration{})
				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.NetworkAndAZ(api.NetworkAndAZConfiguration{})

				Expect(err).To(MatchError("could not make api request to network and AZ endpoint: api endpoint failed"))
			})
		})
	})

	Describe("Properties", func() {
		It("assigns director configuration properties", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := directorService.Properties(api.DirectorProperties{
				IAASConfiguration:     json.RawMessage(`{"prop": "other", "value": "one"}`),
				DirectorConfiguration: json.RawMessage(`{"prop": "blah", "value": "nothing"}`),
				SecurityConfiguration: json.RawMessage(`{"hello": "goodbye"}`),
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/properties"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"iaas_configuration": {"prop": "other", "value": "one"},
				"director_configuration": {"prop": "blah", "value": "nothing"},
				"security_configuration": {"hello": "goodbye"}
			}`))
		})

		Context("when some of the configurations are empty", func() {
			It("returns only configurations that are populated", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := directorService.Properties(api.DirectorProperties{
					IAASConfiguration:     json.RawMessage(`{"prop": "other", "value": "one"}`),
					DirectorConfiguration: json.RawMessage(`{"prop": "blah", "value": "nothing"}`),
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))
				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("PUT"))
				Expect(req.URL.Path).To(Equal("/api/v0/staged/director/properties"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				jsonBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonBody).To(MatchJSON(`{
					"iaas_configuration": {"prop": "other", "value": "one"},
					"director_configuration": {"prop": "blah", "value": "nothing"}
				}`))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := directorService.Properties(api.DirectorProperties{})

				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.Properties(api.DirectorProperties{})

				Expect(err).To(MatchError("could not make api request to director properties: api endpoint failed"))
			})
		})
	})
})
