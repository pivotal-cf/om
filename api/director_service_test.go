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

		client.DoReturns(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)
	})

	Describe("AZConfiguration", func() {
		It("configures availability zones", func() {
			err := directorService.AZConfiguration(api.AZConfiguration{
				AvailabilityZones: json.RawMessage(`[{"az_name": "1"}]`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"availability_zones": [
					{"az_name": "1"}
				]
			}`))
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := directorService.AZConfiguration(api.AZConfiguration{})
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.AZConfiguration(api.AZConfiguration{})

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/availability_zones: api endpoint failed"))
			})
		})
	})

	Describe("NetworksConfiguration", func() {
		It("configures networks", func() {
			err := directorService.NetworksConfiguration(json.RawMessage(`{"networks": [{"network_property": "yup"}]}`))
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"networks": [
					{"network_property": "yup"}
				]
			}`))
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := directorService.NetworksConfiguration(json.RawMessage("{}"))
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.NetworksConfiguration(json.RawMessage("{}"))
				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/networks: api endpoint failed"))
			})
		})
	})

	Describe("NetworkAndAZ", func() {
		It("creates an network and az assignment", func() {
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

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/network_and_az: api endpoint failed"))
			})
		})
	})

	Describe("Properties", func() {
		It("assigns director configuration properties", func() {
			err := directorService.Properties(api.DirectorProperties{
				IAASConfiguration:     json.RawMessage(`{"prop": "other", "value": "one"}`),
				DirectorConfiguration: json.RawMessage(`{"prop": "blah", "value": "nothing"}`),
				SecurityConfiguration: json.RawMessage(`{"hello": "goodbye"}`),
				SyslogConfiguration:   json.RawMessage(`{"imsyslog": "yep"}`),
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
				"security_configuration": {"hello": "goodbye"},
				"syslog_configuration":{"imsyslog": "yep"}
			}`))
		})

		Context("when some of the configurations are empty", func() {
			It("returns only configurations that are populated", func() {
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

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/properties: api endpoint failed"))
			})
		})
	})
})
