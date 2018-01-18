package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureBosh", func() {
	var (
		boshService       *fakes.BoshFormService
		diagnosticService *fakes.DiagnosticService
		logger            *fakes.Logger
	)

	BeforeEach(func() {
		boshService = &fakes.BoshFormService{}
		diagnosticService = &fakes.DiagnosticService{}
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("configures bosh", func() {
			testBoshConfigurationWithAZs(
				boshService,
				diagnosticService,
				logger,
				`{"availability_zones": [{"name": "some-az-name"},{"name": "some-az-other"}]}`,
				"_method=the-rails&authenticity_token=some-auth-token&availability_zones%5Bavailability_zones%5D%5B%5D%5Biaas_identifier%5D=some-az-name&availability_zones%5Bavailability_zones%5D%5B%5D%5Biaas_identifier%5D=some-az-other",
			)
		})

		It("configures bosh for vSphere", func() {
			testBoshConfigurationWithAZs(
				boshService,
				diagnosticService,
				logger,
				`{"availability_zones": [{"name": "some-az-name","cluster": "cluster-1","resource_pool": "pool-1"},{"name": "some-az-other", "cluster": "cluster-2","resource_pool": "pool-2"}]}`,
				"_method=the-rails&authenticity_token=some-auth-token&availability_zones%5Bavailability_zones%5D%5B%5D%5Bcluster%5D=cluster-1&availability_zones%5Bavailability_zones%5D%5B%5D%5Bcluster%5D=cluster-2&availability_zones%5Bavailability_zones%5D%5B%5D%5Bname%5D=some-az-name&availability_zones%5Bavailability_zones%5D%5B%5D%5Bname%5D=some-az-other&availability_zones%5Bavailability_zones%5D%5B%5D%5Bresource_pool%5D=pool-1&availability_zones%5Bavailability_zones%5D%5B%5D%5Bresource_pool%5D=pool-2",
			)
		})

		It("configures bosh with no availability zones", func() {
			command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

			boshService.GetFormReturns(api.Form{
				Action:            "form-action",
				AuthenticityToken: "some-auth-token",
				RailsMethod:       "the-rails",
			}, nil)

			boshService.NetworksReturns(map[string]string{
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
						"opentsdb_ip": "some-hmforwarder-ip",
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
						"subnets": [
							{
								"iaas_identifier": "some-iaas-identifier",
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
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

			format, content = logger.PrintfArgsForCall(3)
			Expect(fmt.Sprintf(format, content...)).To(Equal("configuring network options for bosh tile"))

			format, content = logger.PrintfArgsForCall(4)
			Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

			format, content = logger.PrintfArgsForCall(5)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

			Expect(boshService.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

			Expect(boshService.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
			}))

			Expect(boshService.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

			Expect(boshService.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string&director_configuration%5Bopentsdb_ip%5D=some-hmforwarder-ip",
			}))

			Expect(boshService.GetFormArgsForCall(2)).To(Equal("/infrastructure/security_tokens/edit"))

			Expect(boshService.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
			}))

			Expect(boshService.GetFormArgsForCall(3)).To(Equal("/infrastructure/networks/edit"))

			Expect(boshService.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&infrastructure%5Bicmp_checks_enabled%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bguid%5D=0&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bname%5D=some-network&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bservice_network%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=null-az&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bcidr%5D=10.0.1.0%2F24&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bdns%5D=8.8.8.8&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bgateway%5D=10.0.1.1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Biaas_identifier%5D=some-iaas-identifier&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Breserved_ip_ranges%5D=10.0.1.0-10.0.1.4",
			}))

			Expect(boshService.GetFormArgsForCall(4)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

			Expect(boshService.PostFormArgsForCall(4)).To(Equal(api.PostFormInput{
				Form: api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				},
				EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=null-az",
			}))
		})

		Context("when a network already exists", func() {
			It("does network assignment", func() {
				command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

				boshService.GetFormReturns(api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				}, nil)

				boshService.NetworksReturns(map[string]string{
					"some-network-name":       "some-network-guid",
					"some-other-network-name": "some-other-network-guid",
				}, nil)

				boshService.AvailabilityZonesReturns(map[string]string{
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
				Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

				format, content = logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

				format, content = logger.PrintfArgsForCall(4)
				Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

				Expect(boshService.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

				Expect(boshService.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
				}))

				Expect(boshService.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

				Expect(boshService.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string",
				}))

				Expect(boshService.GetFormArgsForCall(2)).To(Equal("/infrastructure/security_tokens/edit"))

				Expect(boshService.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
				}))

				Expect(boshService.GetFormArgsForCall(3)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

				Expect(boshService.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
					Form: api.Form{
						Action:            "form-action",
						AuthenticityToken: "some-auth-token",
						RailsMethod:       "the-rails",
					},
					EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=guid-1",
				}))

				Expect(boshService.AvailabilityZonesCallCount()).To(Equal(1))
			})
		})

		Context("when the iaas is azure", func() {
			It("does not attempt to get the availability zones", func() {
				command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

				boshService.GetFormReturns(api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				}, nil)

				diagnosticService.ReportReturns(api.DiagnosticReport{
					InfrastructureType: "azure",
				}, nil)

				boshService.NetworksReturns(map[string]string{
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
					"--network-assignment",
					`{
						"singleton_availability_zone": "some-az-name",
					  "network": "some-other-network-name"
					}`,
					"--networks-configuration",
					`{
						"icmp_checks_enabled": false,
						"networks": [
							{
								"name": "opsman-network",
								"subnets": [
									{
										"iaas_identifier": "vpc-subnet-id-1",
										"cidr": "10.0.0.0/24",
										"reserved_ip_ranges": "10.0.0.0-10.0.0.4",
										"dns": "8.8.8.8",
										"gateway": "10.0.0.1",
										"availability_zones": ["some-az1"]
									}
								]
							}
							]
						}`,
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(boshService.AvailabilityZonesCallCount()).To(Equal(0))
			})
		})

		Context("when the network is configured and deployed", func() {
			It("ignores all network changes", func() {
				command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

				boshService.GetFormReturns(api.Form{
					Action:            "form-action",
					AuthenticityToken: "some-auth-token",
					RailsMethod:       "the-rails",
				}, nil)

				diagnosticService.ReportReturns(api.DiagnosticReport{
					DeployedProducts: []api.DiagnosticProduct{
						{
							Name: "p-bosh",
						},
					},
				}, nil)

				boshService.NetworksReturns(map[string]string{
					"some-network-name":       "some-network-guid",
					"some-other-network-name": "some-other-network-guid",
				}, nil)

				boshService.AvailabilityZonesReturns(map[string]string{
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
					"--az-configuration",
					`{
						"availability_zones": [
						{"name": "some-az1"},
						{"name": "some-az2"}
						]
					}`,
					"--networks-configuration",
					`{
						"icmp_checks_enabled": false,
						"networks": [
							{
								"name": "opsman-network",
								"subnets": [
									{
										"iaas_identifier": "vpc-subnet-id-1",
										"cidr": "10.0.0.0/24",
										"reserved_ip_ranges": "10.0.0.0-10.0.0.4",
										"dns": "8.8.8.8",
										"gateway": "10.0.0.1",
										"availability_zones": ["some-az1"]
									}
								]
							}
							]
						}`,
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(boshService.GetFormCallCount()).To(Equal(3))
				Expect(boshService.AvailabilityZonesCallCount()).To(Equal(0))
				Expect(boshService.NetworksCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(3)
				Expect(fmt.Sprintf(format, content...)).To(Equal("skipping network configuration: detected deployed director - cannot modify network"))
			})

			Context("when error occurs", func() {
				Context("when fetching the diagnostic report fails", func() {
					It("returns an error", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

						diagnosticService.ReportReturns(api.DiagnosticReport{}, errors.New("beep boop"))

						err := command.Execute([]string{"--az-configuration", `{"some": "configuration"}`})
						Expect(err).To(MatchError("beep boop"))
					})
				})
			})

			Context("when an empty JSON object is provided", func() {
				AfterEach(func() {
					Expect(boshService.PostFormCallCount()).To(Equal(0), "should not post")

					format, content := logger.PrintfArgsForCall(0)
					Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))
				})

				Context("to the --az-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--az-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --director-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--director-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --iaas-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--iaas-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --network-assignment flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--network-assignment", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --networks-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--networks-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --resource-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--resource-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("to the --security-configuration flag", func() {
					It("is ignored", func() {
						command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
						err := command.Execute([]string{"--security-configuration", "{}"})
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("error cases", func() {
			Context("when no configuration flags are passed", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("at least one configuration flag must be provided. Please see usage for more information."))
				})
			})

			Context("when an invalid flag is passed", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)
					err := command.Execute([]string{"--not-a-real-flag"})
					Expect(err).To(MatchError("flag provided but not defined: -not-a-real-flag"))
				})
			})

			Context("when the form can't be fetched", func() {
				It("returns an error", func() {
					boshService.GetFormReturns(api.Form{}, errors.New("meow meow meow"))
					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

					err := command.Execute([]string{"--iaas-configuration", `{"some": "configuration"}`})
					Expect(err).To(MatchError("could not fetch form: meow meow meow"))
				})
			})

			Context("when the json can't be decoded", func() {
				It("returns an error", func() {
					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

					err := command.Execute([]string{"--iaas-configuration", "%$#@#$"})
					Expect(err).To(MatchError("could not decode json: invalid character '%' looking for beginning of value"))
				})
			})

			Context("when configuring the tile fails", func() {
				It("returns an error", func() {
					boshService.PostFormReturns(errors.New("NOPE"))

					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

					err := command.Execute([]string{"--iaas-configuration", `{"invalid": "configuration"}`})
					Expect(err).To(MatchError("tile failed to configure: NOPE"))
				})
			})

			Context("when retrieving the list of availability zones fails", func() {
				It("returns an error", func() {
					boshService.AvailabilityZonesReturns(map[string]string{}, errors.New("FAIL"))

					diagnosticService.ReportReturns(
						api.DiagnosticReport{}, nil)

					command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

					err := command.Execute([]string{
						"--az-configuration", `{"some": "configuration"}`,
						"--networks-configuration", `{"some": "configuration"}`,
					})
					Expect(err).To(MatchError("could not fetch availability zones: FAIL"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage for the command", func() {
			command := commands.NewConfigureBosh(nil, nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "configures the bosh director that is deployed by the Ops Manager",
				ShortDescription: "configures Ops Manager deployed bosh director",
				Flags:            command.Options,
			}))
		})
	})
})

func testBoshConfigurationWithAZs(boshService *fakes.BoshFormService, diagnosticService *fakes.DiagnosticService, logger *fakes.Logger, azConfiguration, azEncodedPayload string) {
	command := commands.NewConfigureBosh(boshService, diagnosticService, logger)

	boshService.GetFormReturns(api.Form{
		Action:            "form-action",
		AuthenticityToken: "some-auth-token",
		RailsMethod:       "the-rails",
	}, nil)

	boshService.AvailabilityZonesReturns(map[string]string{
		"some-az-name":  "guid-1",
		"some-az-other": "guid-2",
	}, nil)

	boshService.NetworksReturns(map[string]string{
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
					"subnets": [
						{
							"iaas_identifier": "some-iaas-identifier",
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
		"--resource-configuration",
		`{
			"director": {
				"instance_type": {
					"id": "m1.medium"
				},
				"persistent_disk": {
					"size_mb": "20480"
				},
				"internet_connected": true,
				"elb_names": ["my-elb"]
			},
			"compilation": {
				"instances": 1,
				"instance_type": {
					"id": "m1.medium"
				},
				"internet_connected": true,
				"elb_names": ["my-elb"]
			}
		}`,
	})

	Expect(err).NotTo(HaveOccurred())

	format, content := logger.PrintfArgsForCall(0)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring iaas specific options for bosh tile"))

	format, content = logger.PrintfArgsForCall(1)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring director options for bosh tile"))

	format, content = logger.PrintfArgsForCall(2)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring security options for bosh tile"))

	format, content = logger.PrintfArgsForCall(3)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring resources for bosh tile"))

	format, content = logger.PrintfArgsForCall(4)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring availability zones for bosh tile"))

	format, content = logger.PrintfArgsForCall(5)
	Expect(fmt.Sprintf(format, content...)).To(Equal("configuring network options for bosh tile"))

	format, content = logger.PrintfArgsForCall(6)
	Expect(fmt.Sprintf(format, content...)).To(Equal("assigning az and networks for bosh tile"))

	format, content = logger.PrintfArgsForCall(7)
	Expect(fmt.Sprintf(format, content...)).To(Equal("finished configuring bosh tile"))

	Expect(boshService.GetFormArgsForCall(0)).To(Equal("/infrastructure/iaas_configuration/edit"))

	Expect(boshService.PostFormArgsForCall(0)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&iaas_configuration%5Bauth_json%5D=%7B%22some-auth-field%22%3A+%22some-service-key%22%2C%22some-private_key%22%3A+%22some-key%22%7D&iaas_configuration%5Bdefault_deployment_tag%5D=my-vms&iaas_configuration%5Bproject%5D=some-project",
	}))

	Expect(boshService.GetFormArgsForCall(1)).To(Equal("/infrastructure/director_configuration/edit"))

	Expect(boshService.PostFormArgsForCall(1)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&director_configuration%5Bhm_pager_duty_options%5D%5Benabled%5D=true&director_configuration%5Bmetrics_ip%5D=some-metrics-ip&director_configuration%5Bntp_servers_string%5D=some-ntp-servers-string",
	}))

	Expect(boshService.GetFormArgsForCall(2)).To(Equal("/infrastructure/security_tokens/edit"))

	Expect(boshService.PostFormArgsForCall(2)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&security_tokens%5Btrusted_certificates%5D=some-trusted-certificates&security_tokens%5Bvm_password_type%5D=some-vm-password-type",
	}))

	Expect(boshService.GetFormArgsForCall(3)).To(Equal("/infrastructure/director/resources/edit"))

	Expect(boshService.PostFormArgsForCall(3)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&product_resources_form%5Bcompilation%5D%5Belb_names%5D=my-elb&product_resources_form%5Bcompilation%5D%5Binstances%5D=1&product_resources_form%5Bcompilation%5D%5Binternet_connected%5D=true&product_resources_form%5Bcompilation%5D%5Bvm_type_id%5D=m1.medium&product_resources_form%5Bdirector%5D%5Bdisk_type_id%5D=20480&product_resources_form%5Bdirector%5D%5Belb_names%5D=my-elb&product_resources_form%5Bdirector%5D%5Binternet_connected%5D=true&product_resources_form%5Bdirector%5D%5Bvm_type_id%5D=m1.medium",
	}))

	Expect(boshService.GetFormArgsForCall(4)).To(Equal("/infrastructure/availability_zones/edit"))

	Expect(boshService.PostFormArgsForCall(4)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: azEncodedPayload,
	}))

	Expect(boshService.GetFormArgsForCall(5)).To(Equal("/infrastructure/networks/edit"))

	Expect(boshService.PostFormArgsForCall(5)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&infrastructure%5Bicmp_checks_enabled%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bguid%5D=0&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bname%5D=some-network&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bservice_network%5D=1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=guid-1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bavailability_zone_references%5D%5B%5D=guid-2&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bcidr%5D=10.0.1.0%2F24&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bdns%5D=8.8.8.8&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Bgateway%5D=10.0.1.1&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Biaas_identifier%5D=some-iaas-identifier&network_collection%5Bnetworks_attributes%5D%5B0%5D%5Bsubnets%5D%5B0%5D%5Breserved_ip_ranges%5D=10.0.1.0-10.0.1.4",
	}))

	Expect(boshService.GetFormArgsForCall(6)).To(Equal("/infrastructure/director/az_and_network_assignment/edit"))

	Expect(boshService.PostFormArgsForCall(6)).To(Equal(api.PostFormInput{
		Form: api.Form{
			Action:            "form-action",
			AuthenticityToken: "some-auth-token",
			RailsMethod:       "the-rails",
		},
		EncodedPayload: "_method=the-rails&authenticity_token=some-auth-token&bosh_product%5Bnetwork_reference%5D=some-other-network-guid&bosh_product%5Bsingleton_availability_zone_reference%5D=guid-1",
	}))
}
