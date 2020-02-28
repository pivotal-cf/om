package api_test

import (
    "encoding/json"
    "net/http"

    "github.com/onsi/gomega/ghttp"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/pivotal-cf/om/api"
    "github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("Director", func() {
    var (
        server  *ghttp.Server
        stderr  *fakes.Logger
        service api.Api
    )

    BeforeEach(func() {
        server = ghttp.NewServer()

        stderr = &fakes.Logger{}
        service = api.New(api.ApiInput{
            Client: httpClient{server.URL()},
            Logger: stderr,
        })
    })

    AfterEach(func() {
        server.Close()
    })

    Describe("AZConfiguration", func() {
        When("happy path", func() {
            BeforeEach(func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "existing-iaas-guid",
								"name": "existing-iaas"
							}, {
								"guid": "new-iaas-guid",
								"name": "new-iaas"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{
							"availability_zones": [{
								"guid": "existing-az-guid",
								"name": "existing-az",
								"clusters": [{
									"cluster":"pizza",
									"guid":"pepperoni",
									"res_pool":"dcba"
								}]
							}]
						}`),
                    ),
                )
            })

            It("configures availability zones", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/availability_zones/existing-az-guid"),
                        ghttp.VerifyJSON(`{
							"availability_zone": {
								"a_field": "some_val",
								"guid": "existing-az-guid",
								"name": "existing-az",
								"iaas_configuration_guid": "existing-iaas-guid",
								"clusters": [{
									"cluster": "pizza",
									"guid": "pepperoni",
									"res_pool": "abcd"
								}]
							}
						}`),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.VerifyJSON(`{
							"availability_zone":{
								"name": "new-az",
								"iaas_configuration_guid": "new-iaas-guid"
							}
						}`),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                )
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[
						{
							"clusters": [{
								"cluster": "pizza",
								"res_pool": "abcd"
							}],
							"iaas_configuration_name": "existing-iaas",
							"name": "existing-az",
							"a_field":"some_val"
						}, {
							"name": "new-az",
							"iaas_configuration_name": "new-iaas"
						}
					]`),
                }, false)
                Expect(err).ToNot(HaveOccurred())
                Expect(stderr.Invocations()).To(HaveLen(1))
                message := stderr.PrintlnArgsForCall(0)
                Expect(message[0]).To(Equal("successfully fetched AZs, continuing"))
            })

            It("preserves all provided fields", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.VerifyJSON(`{
							"availability_zone":{
								"name": "new-az",
								"iaas_configuration_guid": "new-iaas-guid"
							}
						}`),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                )
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{
						"name": "new-az",
						"iaas_configuration_name": "new-iaas"
					}]`),
                }, false)
                Expect(err).ToNot(HaveOccurred())
            })
        })

        When("the Ops Manager does not support retrieving existing availability zones", func() {
            It("continues to configure the availability zones", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "new-iaas-guid",
								"name": "new-iaas"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusNotFound, ""),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.VerifyJSON(`{
							"availability_zone":{
								"name": "new-az",
								"iaas_configuration_guid": "new-iaas-guid"
							}
						}`),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                )
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new-az", "iaas_configuration_name": "new-iaas"}]`),
                }, false)
                Expect(err).ToNot(HaveOccurred())

                Expect(stderr.PrintlnCallCount()).To(Equal(1))
                warning := stderr.PrintlnArgsForCall(0)
                Expect(warning[0]).To(Equal(
                    "unable to retrieve existing AZ configuration, attempting to configure anyway"))
            })
        })

        When("there is only 1 az config passed in without a name and no iaases exists", func() {
            It("send the request anyway", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": []
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones": [{"name": "existing", "guid":"123"}]}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zone": [{"name": "new"}]}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new", "iaas_configuration_name":"not-existing"}]`),
                }, false)
                Expect(err).ToNot(HaveOccurred())
            })
        })

        When("there is only 1 az config passed in without a name and only 1 iaas config is returned by the api", func() {
            It("send the request anyway", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
						"iaas_configurations": [{
							"guid": "missing-guid",
							"name": "missing"
						}]
					}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones": [{"name": "existing", "guid":"123"}]}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zone": [{"name": "existing"}]}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new"}]`),
                }, false)
                Expect(err).ToNot(HaveOccurred())
            })
        })

        Context("failure cases", func() {
            It("returns an error when the provided AZ config is malformed", func() {
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage("{malformed"),
                }, false)

                Expect(err).To(MatchError(HavePrefix("provided AZ config is not well-formed JSON")))
            })

            It("returns an error when the provided AZ config does not include a name", func() {
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage("[{}]"),
                }, false)

                Expect(err).To(MatchError(HavePrefix("provided AZ config [0] does not specify the AZ 'name'")))
            })

            It("returns an error when the GET iaas_configurations request fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                            server.CloseClientConnections()
                        }),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)
                Expect(err).To(MatchError(ContainSubstring("could not send api request to GET /api/v0/staged/director/iaas_configurations")))
            })

            It("returns an error when the GET availability zones http status is not a 200 or 404", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusInternalServerError, `{}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)
                Expect(err).To(MatchError(HavePrefix("received unexpected status while fetching AZ configuration")))
                Expect(err).To(MatchError(ContainSubstring("500")))
            })

            It("returns an error when the GET returns malformed existing AZs", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, "malformed"),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, false)

                Expect(err).To(MatchError(HavePrefix(
                    "problem retrieving existing AZs: response is not well-formed")))
            })

            It("ignores warnings when the PUT http status is 207 and ignoreVerifierWarnings is true", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones":[]}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{}, true)
                Expect(err).To(BeNil())
            })

            It("returns an error when the PUT and POST http status is 207 and ignoreVerifierWarnings is false", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "existing-guid",
								"name": "existing"
							}, {
								"guid": "new-guid",
								"name": "new"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones": [{"name": "existing", "guid":"123"}]}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new", "iaas_configuration_name": "new"}]`),
                }, false)
                Expect(err).To(MatchError(ContainSubstring("Multi-Status")))

                server.Reset()

                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "existing-guid",
								"name": "existing"
							},{
								"guid": "new-guid",
								"name": "new"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones": [{"name": "existing", "guid":"123"}]}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/availability_zones/123"),
                        ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                    ),
                )

                err = service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "existing", "iaas_configuration_name": "existing"}]`),
                }, false)
                Expect(err).To(MatchError(ContainSubstring("Multi-Status")))
            })

            It("returns an error when there is 1 az config passed in and multiple iaas configs are returned by the api", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "missing-guid",
								"name": "missing"
							}, {
								"guid": "another-guid",
								"name": "another"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones": [{"name": "existing", "guid":"123"}]}`),
                    ),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new", "iaas_configuration_name": "new"}]`),
                }, false)
                Expect(err).To(MatchError(ContainSubstring("provided AZ 'iaas_configuration_name' ('new') doesn't match any existing iaas_configurations")))
            })

            It("returns an error when the PUT http status is non-200", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "new-guid",
								"name": "new"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones":[]}`),
                    ),
                    ghttp.RespondWith(http.StatusInternalServerError, `{}`),
                )
                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new", "iaas_configuration_name": "new"}]`)}, false)
                Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
            })

            It("returns an error when the PUT to the api endpoint fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "new-guid",
								"name": "new"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/availability_zones"),
                        ghttp.RespondWith(http.StatusOK, `{"availability_zones":[]}`),
                    ),
                    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                        server.CloseClientConnections()
                    }),
                )

                err := service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
                    AvailabilityZones: json.RawMessage(`[{"name": "new", "iaas_configuration_name": "new"}]`)}, false)
                Expect(err).To(MatchError(ContainSubstring("could not send api request to POST /api/v0/staged/director/availability_zones")))
            })
        })
    })

    Describe("NetworksConfiguration", func() {
        Context("happy path", func() {
            BeforeEach(func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
                        ghttp.VerifyHeader(map[string][]string{"Content-Type": []string{"application/json"}}),
                        ghttp.RespondWith(http.StatusOK, `{
							"networks": [{
								"guid": "existing-network-guid", 
								"name": "existing-network"
							}]
						}`),
                    ),
                )
            })

            It("configures networks", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/networks"),
                        ghttp.VerifyHeader(map[string][]string{"Content-Type": []string{"application/json"}}),
                        ghttp.VerifyJSON(`{
							"networks": [{
								"name": "yup"
							}]
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{"networks": [{"name": "yup"}]}`),
                })
                Expect(err).ToNot(HaveOccurred())
            })

            It("configures networks and associates existing guids", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/networks"),
                        ghttp.VerifyJSON(`{
							"icmp_checks_enabled":false,
							"networks": [{
								"name": "existing-network",
								"guid": "existing-network-guid"
							}]
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{"icmp_checks_enabled":false, "networks": [{"name":"existing-network"}]}`),
                })
                Expect(err).ToNot(HaveOccurred())
            })

            It("configures networks and associates existing guids and no guid for new network", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/networks"),
                        ghttp.VerifyJSON(`{
							"icmp_checks_enabled":false,
							"networks": [{
								"name": "existing-network",
								"guid": "existing-network-guid"
							}, {
								"name": "new-network"
							}]
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{"icmp_checks_enabled":false, "networks": [{"name":"existing-network"},{"name":"new-network"}]}`),
                })
                Expect(err).ToNot(HaveOccurred())
            })
        })

        Context("failure cases", func() {
            When("the Ops Manager does not support retrieving existing networks", func() {
                It("continues to configure the networks", func() {
                    server.AppendHandlers(
                        ghttp.CombineHandlers(
                            ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
                            ghttp.RespondWith(http.StatusNotFound, ""),
                        ),
                        ghttp.CombineHandlers(
                            ghttp.VerifyRequest("PUT", "/api/v0/staged/director/networks"),
                            ghttp.RespondWith(http.StatusOK, ""),
                        ),
                    )

                    err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                        Networks: json.RawMessage(`{
							"networks": [{
								"name": "new-network"
							}]
						}`),
                    })
                    Expect(err).ToNot(HaveOccurred())

                    Expect(stderr.PrintlnCallCount()).To(Equal(1))
                    warning := stderr.PrintlnArgsForCall(0)
                    Expect(warning[0]).To(Equal("unable to retrieve existing network configuration, attempting to configure anyway"))
                })
            })

            It("returns an error when the provided network config is malformed", func() {
                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage("{malformed"),
                })

                Expect(err).To(MatchError(HavePrefix("provided networks config is not well-formed JSON")))
            })

            It("returns an error when the provided network config does not include a name", func() {
                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{"networks":[{}]}`),
                })
                Expect(err).To(MatchError(HavePrefix("provided networks config [0] does not specify the network 'name'")))
            })

            It("returns an error when the http status is non-200", func() {
                server.AppendHandlers(
                    ghttp.RespondWith(http.StatusInternalServerError, ""),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{}`),
                })
                Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
            })

            It("returns an error when the api endpoint fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
                        ghttp.RespondWith(http.StatusOK, `{"networks": []}`),
                    ),
                    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                        server.CloseClientConnections()
                    }),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{}`),
                })
                Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/staged/director/networks")))
            })

            It("returns an error when the network endpoint status is non-200", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/networks"),
                        ghttp.RespondWith(http.StatusOK, `{"networks": []}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/networks"),
                        ghttp.RespondWith(http.StatusInternalServerError, ""),
                    ),
                )

                err := service.UpdateStagedDirectorNetworks(api.NetworkInput{
                    Networks: json.RawMessage(`{}`),
                })
                Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
            })
        })
    })

    Describe("NetworkAndAZ", func() {
        Context("happy path", func() {
            It("creates an network and az assignment", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials"),
                        ghttp.RespondWith(http.StatusNotFound, ""),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/network_and_az"),
                        ghttp.VerifyJSON(`{
							"network_and_az": {
								"network": {
									"name": "network_name"
								},
								"singleton_availability_zone": {
									"name": "availability_zone_name"
								}
							}
						}`),
                        ghttp.RespondWith(http.StatusOK, `{}`),
                    ),
                )

                err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
                    NetworkAZ: json.RawMessage(`{
						"network": {
							"name": "network_name"
						},
						"singleton_availability_zone": {
							"name": "availability_zone_name"
						}
					}`),
                })

                Expect(err).ToNot(HaveOccurred())
            })

            When("the director has already been deployed", func() {
                It("issues a warning and doesn't configure the endpoint", func() {
                    server.AppendHandlers(
                        ghttp.CombineHandlers(
                            ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials"),
                            ghttp.RespondWith(http.StatusOK, ""),
                        ),
                    )

                    err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
                        NetworkAZ: json.RawMessage(`{
							"network": {
								"name": "network_name"
							},
							"singleton_availability_zone": {
								"name": "availability_zone_name"
							}
						}`),
                    })

                    Expect(err).ToNot(HaveOccurred())

                    Expect(stderr.PrintlnCallCount()).To(Equal(1))
                    warning := stderr.PrintlnArgsForCall(0)
                    Expect(warning[0]).To(Equal("unable to set network assignment for director as it has already been deployed"))
                })
            })
        })

        Context("failure cases", func() {
            It("returns an error when the http status of the network_and_az endpoint is non-200", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials"),
                        ghttp.RespondWith(http.StatusNotFound, ""),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/network_and_az"),
                        ghttp.RespondWith(http.StatusTeapot, ""),
                    ),
                )

                err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{})
                Expect(err).To(MatchError(ContainSubstring("418")))
            })

            It("returns an error when the api endpoint fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/deployed/director/credentials"),
                        ghttp.RespondWith(http.StatusNotFound, ""),
                    ),
                    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                        server.CloseClientConnections()
                    }),
                )

                err := service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{})
                Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/staged/director/network_and_az")))
            })
        })
    })

    Describe("PropertyInputs", func() {
        Context("happy path", func() {
            It("assigns director configuration properties", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
                        ghttp.VerifyJSON(`{
						"iaas_configuration": {"prop": "other", "value": "one"},
						"director_configuration": {"prop": "blah", "value": "nothing"},
						"dns_configuration": {"recurse": "no"},
						"security_configuration": {"hello": "goodbye"},
						"syslog_configuration": {"imsyslog": "yep"}
					}`),
                    ),
                )
                err := service.UpdateStagedDirectorProperties(api.DirectorProperties(`{
				"iaas_configuration": {"prop": "other", "value": "one"},
				"director_configuration": {"prop": "blah", "value": "nothing"},
				"dns_configuration": {"recurse": "no"},
				"security_configuration": {"hello": "goodbye"},
				"syslog_configuration": {"imsyslog": "yep"}
			}`))

                Expect(err).ToNot(HaveOccurred())
            })

            When("some of the configurations are empty", func() {
                It("returns only configurations that are populated", func() {
                    server.AppendHandlers(
                        ghttp.CombineHandlers(
                            ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
                            ghttp.VerifyJSON(`{
								"iaas_configuration": {"prop": "other", "value": "one"},
								"director_configuration": {"prop": "blah", "value": "nothing"}
							}`),
                        ),
                    )
                    err := service.UpdateStagedDirectorProperties(api.DirectorProperties(`{"iaas_configuration": {"prop": "other", "value": "one"}, "director_configuration": {"prop": "blah", "value": "nothing"}}`))

                    Expect(err).ToNot(HaveOccurred())
                })
            })

        })

        Context("failure cases", func() {
            It("returns an error when the http status is non-200", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
                        ghttp.RespondWith(http.StatusTeapot, ""),
                    ),
                )

                err := service.UpdateStagedDirectorProperties(api.DirectorProperties(``))
                Expect(err).To(MatchError(ContainSubstring("418 I'm a teapot")))
            })

            It("returns an error when the api endpoint fails", func() {
                server.Close()

                err := service.UpdateStagedDirectorProperties(api.DirectorProperties(``))
                Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/staged/director/properties")))
            })
        })
    })

    Describe("IAASConfigurations", func() {
        When("given a list of IAAS Configurations", func() {
            It("creates each iaas configuration if they are new", func() {
                server.AppendHandlers(
                    ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
								"name": "one"
							}
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
							"name": "default",
							"guid": "some-guid",
							"vsphere": "something"
							}
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "one"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("creates and updates iaas configurations", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "existing"
							}, {
								"name": "new"
							}]
						}`),
                    ),
                    ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configuration": {
								"name": "new"
							}
						}`),
                    ),
                )
                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing"},{"name":"new"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("updates existing iaas configuration if the name already exists", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "existing"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
							"name": "existing",
							"guid": "some-guid",
							"vsphere": "something"
							}
						}`),
                    ),
                )
                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("updates existing default configuration if the name default is used in the config", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default"
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
							"name": "default",
							"guid": "some-guid",
							"vsphere": "something"
							}
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "default", "vsphere": "something"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("deletes the empty default configuration", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"vcenter_host": null,
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("DELETE", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.RespondWith(http.StatusNoContent, `{}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
								"name": "something-else",
								"vsphere": "something"
							}
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
							"name": "something-else",
							"guid": "some-guid",
							"vsphere": "something"
							}
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "something-else", "vsphere": "something"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("does not delete the default configuration if it is not empty", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"vcenter_host": "example-host",
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
								"name": "something-else",
								"vsphere": "something"
							}
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.VerifyJSON(`{
							"iaas_configuration": {
							"name": "something-else",
							"guid": "some-guid",
							"vsphere": "something"
							}
						}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "something-else", "vsphere": "something"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("returns an error if DELETE to iaas_configurations endpoint fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"vcenter_host": null,
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("DELETE", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                            server.CloseClientConnections()
                        }),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "something-else", "vsphere": "something"}]`), false)
                Expect(err).To(HaveOccurred())
                Expect(err).To(MatchError(ContainSubstring("could not send api request to DELETE /api/v0/staged/director/iaas_configurations/some-guid")))
            })

            It("returns an error if DELETE to iaas_configurations endpoint returns a status code that is not 204 or 501", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"vcenter_host": null,
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("DELETE", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.RespondWith(http.StatusInternalServerError, `{}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "something-else", "vsphere": "something"}]`), false)
                Expect(err).To(HaveOccurred())
                Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
            })

            It("returns an error if GET to iaas_configurations fails", func() {
                server.Close()

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).To(HaveOccurred())
            })

            It("returns an error if POST to iaas_configurations fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, ""),
                    ),
                    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                        server.CloseClientConnections()
                    }),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).To(HaveOccurred())
            })

            It("returns an error if PUT to iaas_configurations fails", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "existing"
							}]
						}`),
                    ),
                    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
                        server.CloseClientConnections()
                    }),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/staged/director/iaas_configurations/some-guid")))
            })

            It("returns an error if the response body from GET iaas_configurations is not JSON", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, "bad payload"),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).To(MatchError(ContainSubstring("failed to unmarshal JSON response from Ops Manager")))
            })

        })

        When("IAASConfigurations DELETE endpoint is not implemented", func() {
            It("allows later logic to handle non-implementation of the multi-iaas API", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "default",
								"vcenter_host": null,
							}]
						}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("DELETE", "/api/v0/staged/director/iaas_configurations/some-guid"),
                        ghttp.RespondWith(http.StatusNotImplemented, ""),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusNotImplemented, ""),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "something-else", "vsphere": "something"}, {"name": "another", "vsphere": "another"}]`), false)
                Expect(err).To(HaveOccurred())
                Expect(err).To(MatchError(ContainSubstring("multiple iaas_configurations are not allowed for your IAAS.")))
                Expect(err).To(MatchError(ContainSubstring("Supported IAASes include: vsphere and openstack.")))
            })
        })

        When("IAASConfigurations POST endpoint is not implemented", func() {
            It("fails if there are multiple configurations defined", func() {
                server.AppendHandlers(
                    ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusNotImplemented, ""),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "config1"}, {"name": "config2"}]`), false)
                Expect(err).To(MatchError(ContainSubstring("multiple iaas_configurations are not allowed for your IAAS.")))
                Expect(err).To(MatchError(ContainSubstring("Supported IAASes include: vsphere and openstack.")))
            })

            It("sets new IAASConfiguration using UpdateStagedDirectorProperties if there is just one IaaS", func() {
                server.AppendHandlers(
                    ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                    ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                    ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties"),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
                        ghttp.VerifyJSON(`{
								"iaas_configuration": {
									"name": "new"
								}
							}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "new"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })

            It("updates IAASConfiguration using UpdateStagedDirectorProperties if there is just one IaaS", func() {
                server.AppendHandlers(
                    ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("POST", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusNotImplemented, ""),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/properties"),
                        ghttp.RespondWith(http.StatusOK, `{
								"iaas_configuration": {
									"guid": "some-guid",
									"name": "existing"
								}
							}`),
                    ),
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("PUT", "/api/v0/staged/director/properties"),
                        ghttp.VerifyJSON(`{
								"iaas_configuration": {
									"name": "existing",
									"guid": "some-guid",
									"other-field": "value"
								}
							}`),
                    ),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "other-field": "value"}]`), false)
                Expect(err).ToNot(HaveOccurred())
            })
        })

        When("ignore warnings is set to true", func() {
            It("ignores warnings when a 207 is returned on the POST", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "existing"
							}]
						}`),
                    ),
                    ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), true)
                Expect(err).ToNot(HaveOccurred())
            })

            It("ignores warnings when a 207 is returned on the POST", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, ""),
                    ),
                    ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), true)
                Expect(err).ToNot(HaveOccurred())
            })
        })

        When("ignore warnings is set to false", func() {
            It("returns an error when a 207 is returned on the POST", func() {
                server.AppendHandlers(
                    ghttp.CombineHandlers(
                        ghttp.VerifyRequest("GET", "/api/v0/staged/director/iaas_configurations"),
                        ghttp.RespondWith(http.StatusOK, `{
							"iaas_configurations": [{
								"guid": "some-guid",
								"name": "existing"
							}]
						}`),
                    ),
                    ghttp.RespondWith(http.StatusMultiStatus, `{}`),
                )

                err := service.UpdateStagedDirectorIAASConfigurations(api.IAASConfigurationsInput(`[{"name": "existing", "vsphere": "something"}]`), false)
                Expect(err).To(HaveOccurred())
            })
        })
    })
})
