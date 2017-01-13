package commands_test

import (
	"errors"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureBosh", func() {
	var (
		service *fakes.BoshFormService
		logger  *fakes.Logger
	)

	BeforeEach(func() {
		service = &fakes.BoshFormService{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("configures the bosh", func() {
			testBoshConfigurationWithAZs(service, logger, `{"availability_zones": [{"name": "some-az-name"},{"name": "some-az-other"}]}`, "_method=the-rails&authenticity_token=some-auth-token&availability_zones%5Bavailability_zones%5D%5B%5D%5Biaas_identifier%5D=some-az-name&availability_zones%5Bavailability_zones%5D%5B%5D%5Biaas_identifier%5D=some-az-other")
		})
		It("configures the bosh for vSphere", func() {
			testBoshConfigurationWithAZs(service, logger, `{"availability_zones": [{"name": "some-az-name","cluster": "cluster-1","resource_pool": "pool-1"},{"name": "some-az-other", "cluster": "cluster-2","resource_pool": "pool-2"}]}`, "_method=the-rails&authenticity_token=some-auth-token&availability_zones%5Bavailability_zones%5D%5B%5D%5Bcluster%5D=cluster-1&availability_zones%5Bavailability_zones%5D%5B%5D%5Bcluster%5D=cluster-2&availability_zones%5Bavailability_zones%5D%5B%5D%5Bname%5D=some-az-name&availability_zones%5Bavailability_zones%5D%5B%5D%5Bname%5D=some-az-other&availability_zones%5Bavailability_zones%5D%5B%5D%5Bresource_pool%5D=pool-1&availability_zones%5Bavailability_zones%5D%5B%5D%5Bresource_pool%5D=pool-2")
		})

		It("configures the bosh with no availability zones", func() {
			command := commands.NewConfigureBosh(service, logger)

			service.GetFormReturns(api.Form{
				Action:            "form-action",
				AuthenticityToken: "some-auth-token",
				RailsMethod:       "the-rails",
			}, nil)

			service.NetworksReturns(map[string]string{
				"some-network-name":       "some-network-guid",
				"some-other-network-name": "some-other-network-guid",
			}, nil)

			err := command.Execute([]string{
				"--iaas-configuration",
				`{
						"project": "some-project",
						"default_deployment_tag": "my-vms",
						"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
			  }`,
				"--director-configuration",
				`{
					  "ntp_servers_string": "some-ntp-servers-string",
						"metrics_ip": "some-metrics-ip",
						"hm_pager_duty_options": {
							"enabled": true
						}
					}
				}`,
				"--security-configuration",
				`{
						"trusted_certificates": "some-trusted-certificates",
						"vm_password_type": "some-vm-password-type"
				}`,
				"--networks-configuration",
				`{
					"icmp_checks_enabled": true,
					"networks": [{
						"name": "some-network",
						"service_network": true,
						"iaas_identifier": "some-iaas-identifier",
						"subnets": [
							{
								"cidr": "10.0.1.0/24",
								"reserved_ip_ranges": "10.0.1.0-10.0.1.4",
								"dns": "8.8.8.8",
								"gateway": "10.0.1.1",
								"availability_zones": []
							}
						]
					}]
				}`,
				"--network-assignment",
				`{"network": "some-other-network-name"}`,
			})

			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring iaas specific options for bosh tile"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring director options for bosh tile"))

			format, content = logger.PrintfArgsForCall(2)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring network options for bosh tile"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

			format, content = logger.PrintfArgsForCall(4)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

			format, content = logger.PrintfArgsForCall(5)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

			Expect(service.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

			Expect(service.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
			}))

			Expect(service.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

			Expect(service.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string",
			}))

			Expect(service.GetFormArgsForCall(2)).To(Equal("/infrastructure/networks/edit"))

			Expect(service.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&infrastructure%5Bicmp_checks_enabled%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bguid%5D=0&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bname%5D=some-network&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bservice_network%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=null-az&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bcidr%5D=10.0.1.0%2F24&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bdns%5D=8.8.8.8&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bgateway%5D=10.0.1.1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Biaas_identifier%5D=some-iaas-identifier&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Breserved_ip_ranges%5D=10.0.1.0-10.0.1.4",
			}))

			Expect(service.GetFormArgsForCall(3)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

			Expect(service.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=null-az",
			}))

			Expect(service.GetFormArgsForCall(4)).To(Equal("/infrastructure/security_tokens/edit"))

			Expect(service.PostFormArgsForCall(4)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
			}))

			Expect(service.AvailabilityZonesCallCount()).To(Equal(0))
		})

		Context("when a network already exists", func() {
			It("does network assignment", func() {
				command := commands.NewConfigureBosh(service, logger)

				service.GetFormReturns(api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				}, nil)

				service.NetworksReturns(map[string]string{
					"some-network-name":       "some-network-guid",
					"some-other-network-name": "some-other-network-guid",
				}, nil)

				service.AvailabilityZonesReturns(map[string]string{
					"some-az-name":  "guid-1",
					"some-az-other": "guid-2",
				}, nil)

				err := command.Execute([]string{
					"--iaas-configuration",
					`{
						"project": "some-project",
						"default_deployment_tag": "my-vms",
						"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
			  }`,
					"--director-configuration",
					`{
					  "ntp_servers_string": "some-ntp-servers-string",
						"metrics_ip": "some-metrics-ip",
						"hm_pager_duty_options": {
							"enabled": true
						}
					}
				}`,
					"--security-configuration",
					`{
						"trusted_certificates": "some-trusted-certificates",
						"vm_password_type": "some-vm-password-type"
				}`,
					"--network-assignment",
					`{
						"singleton_availability_zone": "some-az-name",
					  "network": "some-other-network-name"
					}`,
				})

				Expect(err).NotTo(HaveOccurred())

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring iaas specific options for bosh tile"))

				format, content = logger.PrintfArgsForCall(1)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring director options for bosh tile"))

				format, content = logger.PrintfArgsForCall(2)
				Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

				format, content = logger.PrintfArgsForCall(4)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

				Expect(service.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

				Expect(service.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
				}))

				Expect(service.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

				Expect(service.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string",
				}))

				Expect(service.GetFormArgsForCall(2)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

				Expect(service.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=guid-1",
				}))

				Expect(service.GetFormArgsForCall(3)).To(Equal("/infrastructure/security_tokens/edit"))

				Expect(service.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
				}))

				Expect(service.AvailabilityZonesCallCount()).To(Equal(1))
			})
		})
		Context("error cases", func() {
			Context("when an invalid flag is passed", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--not-a-real-flag"})
					Expect(err).To(MatchError("flag provided but not defined: -not-a-real-flag"))
				})
			})

			Context("when the form can't be fetched", func() {
				It("returns an error", func() {
					service.GetFormReturns(api.Form{}, errors.New("meow meow meow"))

					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--iaas-configuration", `{}`})
					Expect(err).To(MatchError("could not fetch form: meow meow meow"))
				})
			})

			Context("when the json can't be decoded", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--iaas-configuration", "%$#@#$"})
					Expect(err).To(MatchError("could not decode json: invalid character '%' looking for beginning of value"))
				})
			})

			Context("when configuring the tile fails", func() {
				It("returns an error", func() {
					service.PostFormReturns(errors.New("NOPE"))

					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--iaas-configuration", `{}`})
					Expect(err).To(MatchError("tile failed to configure: NOPE"))
				})
			})

			Context("when retrieving the list of availability zones fails", func() {
				It("returns an error", func() {
					service.AvailabilityZonesReturns(map[string]string{}, errors.New("FAIL"))

					command := commands.NewConfigureBosh(service, logger)

					err := command.Execute([]string{"--az-configuration", `{}`, "--networks-configuration", `{}`})
					Expect(err).To(MatchError("could not fetch availability zones: FAIL"))
				})
			})
		})
	})

	Describe("NetworksConfiguration", func() {
		Describe("EncodeValues", func() {
			It("turns the network configuration into urlencoded form values", func() {
				n := commands.NetworksConfiguration{
					ICMP: true,
					Networks: []commands.NetworkConfiguration{
						{
							Name:           "foo",
							ServiceNetwork: true,
							IAASIdentifier: "something",
							Subnets: []commands.Subnet{
								{
									CIDR:                  "some-cidr",
									ReservedIPRanges:      "reserved-ips",
									DNS:                   "some-dns",
									Gateway:               "some-gateway",
									AvailabilityZoneGUIDs: []string{"one", "two"},
								},
							},
						},
					},
				}

				values, err := query.Values(n)
				Expect(err).NotTo(HaveOccurred())

				Expect(values).To(HaveKeyWithValue("infrastructure[icmp_checks_enabled]", []string{"1"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][name]", []string{"foo"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][service_network]", []string{"1"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][iaas_identifier]", []string{"something"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][cidr]", []string{"some-cidr"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][reserved_ip_ranges]", []string{"reserved-ips"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][dns]", []string{"some-dns"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][gateway]", []string{"some-gateway"}))
				Expect(values).To(HaveKeyWithValue("network_collection[networks_attributes][0][subnets][0][availability_zone_references][]", []string{"one", "two"}))
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage for the command", func() {
			command := commands.NewConfigureBosh(nil, nil)

			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "configures the bosh director that is deployed by the Ops Manager",
				ShortDescription: "configures Ops Manager deployed bosh director",
				Flags:            command.Options,
			}))
		})
	})
})

func testBoshConfigurationWithAZs(service *fakes.BoshFormService, logger *fakes.Logger, azConfiguration, azEncodedPayload string) {

	command := commands.NewConfigureBosh(service, logger)

	service.GetFormReturns(api.Form{
		Action:            "form-action",
		AuthenticityToken: "some-auth-token",
		RailsMethod:       "the-rails",
	}, nil)

	service.AvailabilityZonesReturns(map[string]string{
		"some-az-name":  "guid-1",
		"some-az-other": "guid-2",
	}, nil)

	service.NetworksReturns(map[string]string{
		"some-network-name":       "some-network-guid",
		"some-other-network-name": "some-other-network-guid",
	}, nil)

	err := command.Execute([]string{
		"--iaas-configuration",
		`{
					"project": "some-project",
					"default_deployment_tag": "my-vms",
					"auth_json": "{\"some-auth-field\": \"some-service-key\",\"some-private_key\": \"some-key\"}"
			}`,
		"--director-configuration",
		`{
					"ntp_servers_string": "some-ntp-servers-string",
					"metrics_ip": "some-metrics-ip",
					"hm_pager_duty_options": {
						"enabled": true
					}
				}
			}`,
		"--security-configuration",
		`{
					"trusted_certificates": "some-trusted-certificates",
					"vm_password_type": "some-vm-password-type"
			}`,
		"--az-configuration",
		azConfiguration,
		"--networks-configuration",
		`{
				"icmp_checks_enabled": true,
				"networks": [{
					"name": "some-network",
					"service_network": true,
					"iaas_identifier": "some-iaas-identifier",
					"subnets": [
						{
							"cidr": "10.0.1.0/24",
							"reserved_ip_ranges": "10.0.1.0-10.0.1.4",
							"dns": "8.8.8.8",
							"gateway": "10.0.1.1",
							"availability_zones": [
								"some-az-name",
								"some-az-other"
							]
						}
					]
				}]
			}`,
		"--network-assignment",
		`{"singleton_availability_zone": "some-az-name", "network": "some-other-network-name"}`,
	})

	Expect(err).NotTo(HaveOccurred())

	format, content := logger.PrintfArgsForCall(0)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring iaas specific options for bosh tile"))

	format, content = logger.PrintfArgsForCall(1)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring director options for bosh tile"))

	format, content = logger.PrintfArgsForCall(2)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring availability zones for bosh tile"))

	format, content = logger.PrintfArgsForCall(3)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring network options for bosh tile"))

	format, content = logger.PrintfArgsForCall(4)
	Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

	format, content = logger.PrintfArgsForCall(5)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

	format, content = logger.PrintfArgsForCall(6)
	Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

	Expect(service.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

	Expect(service.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
	}))

	Expect(service.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

	Expect(service.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string",
	}))

	Expect(service.GetFormArgsForCall(2)).To(Equal("/infrastructure/availability_zones/edit"))

	Expect(service.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: azEncodedPayload,
	}))

	Expect(service.GetFormArgsForCall(3)).To(Equal("/infrastructure/networks/edit"))

	Expect(service.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&infrastructure%5Bicmp_checks_enabled%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bguid%5D=0&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bname%5D=some-network&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bservice_network%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=guid-1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=guid-2&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bcidr%5D=10.0.1.0%2F24&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bdns%5D=8.8.8.8&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bgateway%5D=10.0.1.1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Biaas_identifier%5D=some-iaas-identifier&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Breserved_ip_ranges%5D=10.0.1.0-10.0.1.4",
	}))

	Expect(service.GetFormArgsForCall(4)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

	Expect(service.PostFormArgsForCall(4)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=guid-1",
	}))

	Expect(service.GetFormArgsForCall(5)).To(Equal("/infrastructure/security_tokens/edit"))

	Expect(service.PostFormArgsForCall(5)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
	}))
}
