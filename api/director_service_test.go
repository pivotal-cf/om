package api_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("Director", func() {
	var (
		client  *fakes.HttpClient
		stderr  *fakes.Logger
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		stderr = &fakes.Logger{}
		service = api.New(api.ApiInput{
			Client: client,
			Logger: stderr,
		})
	})

	Describe("AZConfiguration", func() {

		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(strings.NewReader(
							`{"availability_zones": [
									{"guid": "existing-az-guid",
									 "name": "existing-az",
									 "clusters":
										[{"cluster":"pizza",
                                          "guid":"pepperoni",
                                          "res_pool":"dcba"}]}]}`,
						))}, nil
				} else {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}
			}
		})

		It("configures availability zones", func() {
			err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[
          			{"clusters":[{"cluster":"pizza","res_pool":"abcd"}],"name":"existing-az","a_field":"some_val"},
          			{"name": "new-az"}
          			  ]`)}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(stderr.Invocations()).To(HaveLen(1))
			message := stderr.PrintlnArgsForCall(0)
			Expect(message[0]).To(Equal("successfully fetched AZs, continuing"))

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
        		 {"a_field":"some_val","guid": "existing-az-guid","name":"existing-az",
        		     "clusters":[{"cluster":"pizza","guid":"pepperoni","res_pool":"abcd"}]},
        		 {"name": "new-az"}
        		]}`))
		})

		It("preserves all provided fields", func() {
			err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[
          {
            "name": "some-az"
          }
        ]`),
			}, false)
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
            "name": "some-az"
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
						Body:       ioutil.NopCloser(strings.NewReader("some error")),
					}, nil
				}
			})

			It("continues to configure the availability zones", func() {
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage(`[
          {"name": "new-az"}
        ]`),
				}, false)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(2))

				putReq := client.DoArgsForCall(1)

				Expect(putReq.Method).To(Equal("PUT"))
				Expect(putReq.URL.Path).To(Equal("/api/v0/staged/director/availability_zones"))
			})

			It("prints a warning to the operator", func() {
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage(`[
          {"name": "new-az"}
        ]`),
				}, false)
				Expect(err).NotTo(HaveOccurred())

				Expect(stderr.PrintlnCallCount()).To(Equal(1))
				warning := stderr.PrintlnArgsForCall(0)
				Expect(warning[0]).To(Equal(
					"unable to retrieve existing AZ configuration, attempting to configure anyway"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the provided AZ config is malformed", func() {
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage("{malformed"),
				}, false)
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided AZ config is not well-formed JSON")))
			})

			It("returns an error when the provided AZ config does not include a name", func() {
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
					AvailabilityZones: json.RawMessage("[{}]"),
				}, false)
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided AZ config [0] does not specify the AZ 'name'")))
			})

			It("returns an error when the GET http status is not a 200 or 404", func() {
				client.DoReturns(
					&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil,
				)
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)
				Expect(err).To(MatchError(HavePrefix("received unexpected status while fetching AZ configuration")))
				Expect(err).To(MatchError(ContainSubstring("500")))
			})

			It("returns an error when the GET to the api endpoint fails", func() {
				client.DoReturns(
					&http.Response{}, errors.New("api endpoint failed"),
				)

				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)

				Expect(err).To(MatchError(ContainSubstring(
					"could not send api request to GET /api/v0/staged/director/availability_zones: api endpoint failed")))
			})

			It("returns an error when the GET returns malformed existing AZs", func() {
				client.DoReturns(
					&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`malformed`))}, nil,
				)

				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)

				Expect(err).To(MatchError(HavePrefix(
					"problem retrieving existing AZs: response is not well-formed")))
			})

			It("ignores warnings when the PUT http status is 207 and ignoreVerifierWarnings is true", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{"availability_zones": []}`))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusMultiStatus,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, true)
				Expect(err).To(BeNil())
			})

			It("returns an error when the PUT http status is 207 and ignoreVerifierWarnings is false", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{"availability_zones": []}`))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusMultiStatus,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)
				Expect(err).To(MatchError(ContainSubstring("Multi-Status")))
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
				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)
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

				err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/availability_zones: api endpoint failed"))
			})
		})
	})

	Describe("NetworksConfiguration", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(strings.NewReader(
							`{"networks": [{"guid": "existing-network-guid", "name": "existing-network"}]}`,
						))}, nil
				} else {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
				}
			}
		})

		It("configures networks", func() {
			err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
				Networks: json.RawMessage(`{"networks": [{"name": "yup"}]}`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			getReq := client.DoArgsForCall(0)

			Expect(getReq.Method).To(Equal("GET"))
			Expect(getReq.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(getReq.Header.Get("Content-Type")).To(Equal("application/json"))

			req := client.DoArgsForCall(1)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"networks": [
					{"name": "yup"}
				]
			}`))
		})

		It("configures networks and associates existing guids", func() {
			err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
				Networks: json.RawMessage(`{"icmp_checks_enabled":false, "networks": [{"name":"existing-network"}]}`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			getReq := client.DoArgsForCall(0)

			Expect(getReq.Method).To(Equal("GET"))
			Expect(getReq.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(getReq.Header.Get("Content-Type")).To(Equal("application/json"))

			req := client.DoArgsForCall(1)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"icmp_checks_enabled":false,
				"networks": [
					{
						"name": "existing-network",
						"guid": "existing-network-guid"
					}
				]
			}`))
		})

		It("configures networks and associates existing guids and no guid for new network", func() {
			err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
				Networks: json.RawMessage(`{"icmp_checks_enabled":false, "networks": [{"name":"existing-network"},{"name":"new-network"}]}`),
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			getReq := client.DoArgsForCall(0)

			Expect(getReq.Method).To(Equal("GET"))
			Expect(getReq.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(getReq.Header.Get("Content-Type")).To(Equal("application/json"))

			req := client.DoArgsForCall(1)

			Expect(req.Method).To(Equal("PUT"))
			Expect(req.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			jsonBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBody).To(MatchJSON(`{
				"icmp_checks_enabled":false,
				"networks": [
					{
						"name": "existing-network",
						"guid": "existing-network-guid"
					},
					{
						"name": "new-network"
					}
				]
			}`))
		})

		Context("when the Ops Manager does not support retrieving existing networks", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					statusCode := http.StatusOK
					if req.Method == "GET" {
						statusCode = http.StatusNotFound
					}
					return &http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader(`{"errors": "some error"}`)),
					}, nil
				}
			})

			It("continues to configure the networks", func() {
				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage(`{"networks": [
          {"name": "new-network"}
        ]}`),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(2))

				putReq := client.DoArgsForCall(1)

				Expect(putReq.Method).To(Equal("PUT"))
				Expect(putReq.URL.Path).To(Equal("/api/v0/staged/director/networks"))
			})

			It("prints a warning to the operator", func() {
				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage(`{"networks":[
          {"name": "new-network"}
        ]}`),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stderr.PrintlnCallCount()).To(Equal(1))
				warning := stderr.PrintlnArgsForCall(0)
				Expect(warning[0]).To(Equal(
					"unable to retrieve existing network configuration, attempting to configure anyway"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the provided network config is malformed", func() {
				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage("{malformed"),
				})
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided networks config is not well-formed JSON")))
			})

			It("returns an error when the provided network config does not include a name", func() {
				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage(`{"networks":[{}]}`),
				})
				Expect(client.DoCallCount()).To(Equal(0))
				Expect(err).To(MatchError(HavePrefix("provided networks config [0] does not specify the network 'name'")))
			})

			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage("{}"),
				})
				Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{"networks": []}`))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed")
					}
				}

				err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
					Networks: json.RawMessage("{}"),
				})
				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/networks: api endpoint failed"))
			})

			Context("when the network endpoint status is non-200", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						switch req.Method {
						case "GET":
							return &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(strings.NewReader(`{"networks": []}`)),
							}, nil
						case "PUT":
							return &http.Response{
								StatusCode: http.StatusInternalServerError,
								Body:       ioutil.NopCloser(strings.NewReader(``)),
							}, nil
						default:
							return nil, errors.New("unexected method in test")
						}
					}

					err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
						Networks: json.RawMessage("{}"),
					})
					Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
				})
			})
		})
	})

	Describe("NetworkAndAZ", func() {
		It("creates an network and az assignment", func() {
			client.DoReturnsOnCall(0, &http.Response{StatusCode: http.StatusNotFound}, nil)
			client.DoReturnsOnCall(1, &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			}, nil)

			err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
				NetworkAZ: json.RawMessage(`{
					"network": {"name": "network_name"},
					"singleton_availability_zone": {"name": "availability_zone_name"}
				}`),
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))
			req := client.DoArgsForCall(1)

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

		Context("when the director has already been deployed", func() {
			It("issues a warning and doesn't configure the endpoint", func() {
				client.DoReturnsOnCall(0, &http.Response{StatusCode: http.StatusOK}, nil)

				err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
					NetworkAZ: json.RawMessage(`{
					"network": {"name": "network_name"},
					"singleton_availability_zone": {"name": "availability_zone_name"}
				}`),
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))
				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/deployed/director/credentials"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the http status is non-200", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

				err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{})
				Expect(err).To(MatchError(ContainSubstring("418")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturnsOnCall(0, &http.Response{StatusCode: http.StatusNotFound}, nil)
				client.DoReturnsOnCall(1, &http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{})

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/network_and_az: api endpoint failed"))
			})
		})
	})

	Describe("Properties", func() {
		BeforeEach(func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)
		})

		It("assigns director configuration properties", func() {
			err := service.UpdateStagedDirectorProperties(api.DirectorProperties(`{
				"iaas_configuration": {"prop": "other", "value": "one"},
				"director_configuration": {"prop": "blah", "value": "nothing"},
                "dns_configuration": {"recurse": "no"},
				"security_configuration": {"hello": "goodbye"},
				"syslog_configuration":{"imsyslog": "yep"}
			}`))

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
                "dns_configuration": {"recurse": "no"},
				"security_configuration": {"hello": "goodbye"},
				"syslog_configuration":{"imsyslog": "yep"}
			}`))
		})

		Context("when some of the configurations are empty", func() {
			It("returns only configurations that are populated", func() {
				err := service.UpdateStagedDirectorProperties(api.DirectorProperties(`{"iaas_configuration": {"prop": "other", "value": "one"},"director_configuration": {"prop": "blah", "value": "nothing"}}`))

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

				err := service.UpdateStagedDirectorProperties(api.DirectorProperties(``))

				Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
			})

			It("returns an error when the api endpoint fails", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

				err := service.UpdateStagedDirectorProperties(api.DirectorProperties(``))

				Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/director/properties: api endpoint failed"))
			})
		})
	})

	Describe("IAASConfigurations", func() {
		When("given a list of IAAS Configurations", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(strings.NewReader(
								`{"iaas_configurations": [
									{"guid": "some-guid",
									 "name": "existing"}]}`,
							))}, nil
					} else {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}
			})

			It("creates each iaas configuration if they are new", func() {
				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "one"}]`))
				Expect(err).NotTo(HaveOccurred())

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/staged/director/iaas_configurations"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				req = client.DoArgsForCall(1)

				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.Path).To(Equal("/api/v0/staged/director/iaas_configurations"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				jsonBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonBody).To(MatchJSON(`{
				"iaas_configuration": {
					"name": "one"
				}}`))
			})

			It("updates existing iaas configuration if the name already exists", func() {
				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`))
				Expect(err).NotTo(HaveOccurred())

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/staged/director/iaas_configurations"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				req = client.DoArgsForCall(1)

				Expect(req.Method).To(Equal("PUT"))
				Expect(req.URL.Path).To(Equal("/api/v0/staged/director/iaas_configurations/some-guid"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				jsonBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonBody).To(MatchJSON(`{
				"iaas_configuration": {
					"name": "existing",
                    "guid": "some-guid",
                    "vsphere": "something"
				}}`))
			})
		})

		Context("failure cases", func() {
			It("returns error if GET to iaas_configurations fails", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "GET" {
						return nil, fmt.Errorf("error")
					} else {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}

				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error"))
			})

			It("returns error if POST to iaas_configurations fails", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "POST" {
						return nil, fmt.Errorf("error")
					} else {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
					}
				}

				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error"))
			})

			It("returns error if PUT to iaas_configurations fails", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					if req.Method == "PUT" {
						return nil, fmt.Errorf("error")
					} else if req.Method == "GET" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(strings.NewReader(
								`{"iaas_configurations": [
									{"guid": "some-guid",
									 "name": "existing"}]}`,
							))}, nil
					}
					return nil, nil
				}

				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error"))
			})

			It("returns an error if the response body is not JSON", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`bad payload`))}, nil
				}

				err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal JSON response from Ops Manager"))
			})
		})
	})
})
