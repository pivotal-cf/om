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

		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(strings.NewReader(
							`{"availability_zones": [{"guid": "existing-az-guid", "name": "existing-az"}]}`,
						))}, nil
				} else {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}
			}
		})

		It("configures availability zones", func() {
			err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[
          {"name": "existing-az"},
          {"name": "new-az"}
        ]`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			getReq := client.DoArgsForCall(0)
			Expect(getReq.Method).To(Equal("GET"))
			Expect(getReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			Expect(getReq.Header.Get("Content-Type")).To(Equal("application/json"))

			putReq := client.DoArgsForCall(1)

			Expect(putReq.Method).To(Equal("PUT"))
			Expect(putReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			Expect(putReq.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(putReq.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
        "availability_zones": [
         {"guid": "existing-az-guid", "name": "existing-az"},
         {"name": "new-az"}
        ]
			}`))
		})

		It("preserves all provided fields", func() {
			err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[
          {
            "name": "some-az",
            "clusters": [
              {
                "cluster": "some-cluster",
                "resource_pool": "some-resource-pool"
              }
            ]
          }
        ]`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			getReq := client.DoArgsForCall(0)
			Expect(getReq.Method).To(Equal("GET"))
			Expect(getReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			Expect(getReq.Header.Get("Content-Type")).To(Equal("application/json"))

			putReq := client.DoArgsForCall(1)

			Expect(putReq.Method).To(Equal("PUT"))
			Expect(putReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			Expect(putReq.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(putReq.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
        "availability_zones": [
          {
            "name": "some-az",
            "clusters": [
              {
                "cluster": "some-cluster",
                "resource_pool": "some-resource-pool"
              }
            ]
          }
        ]
			}`))
		})

		Context("when the Ops Manager does not support retrieving existing availability zones", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					statusCode := http.StatusOK
					if req.Method == "GET" {
						statusCode = http.StatusNotFound
					}
					return &http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil
				}
			})

			It("continues to configure the availability zones", func() {
				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage(`[
          {"name": "new-az"}
        ]`),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(2))

				putReq := client.DoArgsForCall(1)

				Expect(putReq.Method).To(Equal("PUT"))
				Expect(putReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			})
		})

		Context("failure cases", func() {

			It("returns an error when the provided AZ config is malformed", func() {
				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage("{malformed"),
				})
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided AZ config is not well-formed JSON")))
			})

			It("returns an error when the provided AZ config does not include a name", func() {
				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage("[{}]"),
				})
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided AZ config [0] does not specify the AZ 'name'")))
			})

			It("returns an error when the GET http status is not a 200 or 404", func() {
				client.DoReturns(
					&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil,
				)
				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{})
				Expect(err).To(MatchError(HavePrefix("unable to fetch existing AZ configuration")))
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the GET to the api endpoint fails", func() {
				client.DoReturns(
					&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"),
				)

				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{})

				Expect(err).To(MatchError(HavePrefix("unable to fetch existing AZ configuration")))
				Expect(err).To(MatchError(ContainSubstring(
					"could not send api request to GET /api/v0/staged/director/availability_zones: api endpoint failed")))
			})

			It("returns an error when the GET returns malformed existing AZs", func() {
				client.DoReturns(
					&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`malformed`))}, nil,
				)

				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{})

				Expect(err).To(MatchError(HavePrefix(
					"problem retrieving existing AZs: response is not well-formed")))
			})

			It("returns an error when the PUT http status is non-200", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{"availability_zones": []}`))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}
				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{})
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the PUT to the api endpoint fails", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{"availability_zones": []}`))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed")
					}
				}

				err := directorService.SetAZConfiguration(api.AvailabilityZoneInput{})

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/availability_zones: api endpoint failed"))
			})
		})
	})

	Describe("NetworksConfiguration", func() {
		It("configures networks", func() {
			err := directorService.SetNetworksConfiguration(json.RawMessage(`{"networks": [{"network_property": "yup"}]}`))
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

				err := directorService.SetNetworksConfiguration(json.RawMessage("{}"))
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.SetNetworksConfiguration(json.RawMessage("{}"))
				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/networks: api endpoint failed"))
			})
		})
	})

	Describe("NetworkAndAZ", func() {
		It("creates an network and az assignment", func() {
			err := directorService.SetNetworkAndAZ(api.NetworkAndAZConfiguration{
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

				err := directorService.SetNetworkAndAZ(api.NetworkAndAZConfiguration{})
				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.SetNetworkAndAZ(api.NetworkAndAZConfiguration{})

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/network_and_az: api endpoint failed"))
			})
		})
	})

	Describe("Properties", func() {
		It("assigns director configuration properties", func() {
			err := directorService.SetProperties(api.DirectorProperties{
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
				err := directorService.SetProperties(api.DirectorProperties{
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

				err := directorService.SetProperties(api.DirectorProperties{})

				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := directorService.SetProperties(api.DirectorProperties{})

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/properties: api endpoint failed"))
			})
		})
	})
})
