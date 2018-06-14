package commands_test

import (
	"encoding/json"
	"errors"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigureDirector", func() {
	var (
		logger  *fakes.Logger
		service *fakes.ConfigureDirectorService
		command commands.ConfigureDirector

		azConfiguration       []interface{}
		networkAssignment     map[string]interface{}
		iaasConfiguration     map[string]interface{}
		syslogConfiguration   map[string]interface{}
		networksConfiguration map[string]interface{}
		directorConfiguration map[string]interface{}
		securityConfiguration map[string]interface{}
		resourceConfiguration map[string]interface{}

		templateNetworkAssign map[string]interface{}

		azConfigurationJSON       string
		networkAssignmentJSON     string
		iaasConfigurationJSON     string
		syslogConfigurationJSON   string
		networksConfigurationJSON string
		directorConfigurationJSON string
		securityConfigurationJSON string
		resourceConfigurationJSON string

		configurationMAP          map[string]interface{}
		completeConfigurationJSON []byte
		templateConfigurationJSON []byte
	)

	BeforeEach(func() {
		service = &fakes.ConfigureDirectorService{}
		logger = &fakes.Logger{}
		service.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{
				GUID: "p-bosh-guid",
			},
		}, nil)
		service.ListStagedProductJobsReturns(map[string]string{
			"resource": "some-resource-guid",
		}, nil)
		service.GetStagedProductJobResourceConfigReturns(api.JobProperties{
			InstanceType: api.InstanceType{
				ID: "automatic",
			},
			FloatingIPs: "1.2.3.4",
		}, nil)

		command = commands.NewConfigureDirector(service, logger)

		azConfiguration = []interface{}{map[string]interface{}{
			"name": "AZ1",
			"clusters": []interface{}{map[string]interface{}{
				"cluster": "pizza-boxes",
			}},
		}}
		iaasConfiguration = map[string]interface{}{
			"some-iaas-assignment": "iaas",
		}
		networkAssignment = map[string]interface{}{
			"network": map[string]interface{}{
				"name": "network",
			},
			"singleton_availability_zone": map[string]interface{}{
				"name": "singleton",
			},
		}
		syslogConfiguration = map[string]interface{}{
			"some-syslog-assignment": "syslog",
		}
		networksConfiguration = map[string]interface{}{
			"network": "network-1",
		}
		directorConfiguration = map[string]interface{}{
			"some-director-assignment": "director",
		}
		securityConfiguration = map[string]interface{}{
			"some-security-assignment": "security",
		}
		resourceConfiguration = map[string]interface{}{
			"resource": map[string]interface{}{
				"instance_type": map[string]interface{}{
					"id": "some-type",
				},
			},
		}

		templateNetworkAssign = map[string]interface{}{
			"network": map[string]interface{}{
				"name": "((network_name))",
			},
			"singleton_availability_zone": map[string]interface{}{
				"name": "singleton",
			},
		}

		azConfigurationByte, err := json.Marshal(azConfiguration)
		iaasConfigurationByte, err := json.Marshal(iaasConfiguration)
		networkAssignmentByte, err := json.Marshal(networkAssignment)
		syslogConfigurationByte, err := json.Marshal(syslogConfiguration)
		networksConfigurationByte, err := json.Marshal(networksConfiguration)
		directorConfigurationByte, err := json.Marshal(directorConfiguration)
		securityConfigurationByte, err := json.Marshal(securityConfiguration)
		resourceConfigurationByte, err := json.Marshal(resourceConfiguration)

		azConfigurationJSON = string(azConfigurationByte)
		iaasConfigurationJSON = string(iaasConfigurationByte)
		networkAssignmentJSON = string(networkAssignmentByte)
		syslogConfigurationJSON = string(syslogConfigurationByte)
		networksConfigurationJSON = string(networksConfigurationByte)
		directorConfigurationJSON = string(directorConfigurationByte)
		securityConfigurationJSON = string(securityConfigurationByte)
		resourceConfigurationJSON = string(resourceConfigurationByte)

		configurationMAP = map[string]interface{}{}
		configurationMAP["network-assignment"] = networkAssignment
		configurationMAP["az-configuration"] = azConfiguration
		configurationMAP["networks-configuration"] = networksConfiguration
		configurationMAP["director-configuration"] = directorConfiguration
		configurationMAP["iaas-configuration"] = iaasConfiguration
		configurationMAP["security-configuration"] = securityConfiguration
		configurationMAP["syslog-configuration"] = syslogConfiguration
		configurationMAP["resource-configuration"] = resourceConfiguration

		completeConfigurationJSON, err = json.Marshal(configurationMAP)
		Expect(err).NotTo(HaveOccurred())

		configurationMAP["network-assignment"] = templateNetworkAssign
		templateConfigurationJSON, err = json.Marshal(configurationMAP)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Execute", func() {
		ExpectDirectorToBeConfiguredCorrectly := func() {
			Expect(service.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(1))
			Expect(service.UpdateStagedDirectorAvailabilityZonesArgsForCall(0)).To(Equal(api.AvailabilityZoneInput{
				AvailabilityZones: json.RawMessage(`[{"clusters":[{"cluster":"pizza-boxes"}],"name":"AZ1"}]`),
			}))
			Expect(service.UpdateStagedDirectorNetworksCallCount()).To(Equal(1))
			Expect(service.UpdateStagedDirectorNetworksArgsForCall(0)).To(Equal(api.NetworkInput{
				Networks: json.RawMessage(`{"network":"network-1"}`),
			}))

			Expect(service.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(1))
			Expect(service.UpdateStagedDirectorNetworkAndAZArgsForCall(0)).To(Equal(api.NetworkAndAZConfiguration{
				NetworkAZ: json.RawMessage(`{"network":{"name":"network"},"singleton_availability_zone":{"name":"singleton"}}`),
			}))
			Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
			Expect(service.UpdateStagedDirectorPropertiesArgsForCall(0)).To(Equal(api.DirectorProperties{
				DirectorConfiguration: json.RawMessage(`{"some-director-assignment":"director"}`),
				IAASConfiguration:     json.RawMessage(`{"some-iaas-assignment":"iaas"}`),
				SecurityConfiguration: json.RawMessage(`{"some-security-assignment":"security"}`),
				SyslogConfiguration:   json.RawMessage(`{"some-syslog-assignment":"syslog"}`),
			}))
			Expect(service.GetStagedProductByNameCallCount()).To(Equal(1))
			Expect(service.GetStagedProductByNameArgsForCall(0)).To(Equal("p-bosh"))
			Expect(service.ListStagedProductJobsCallCount()).To(Equal(1))
			Expect(service.ListStagedProductJobsArgsForCall(0)).To(Equal("p-bosh-guid"))
			Expect(service.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGUID, instanceGroupGUID := service.GetStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGUID).To(Equal("p-bosh-guid"))
			Expect(instanceGroupGUID).To(Equal("some-resource-guid"))
			Expect(service.UpdateStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGUID, instanceGroupGUID, jobConfiguration := service.UpdateStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGUID).To(Equal("p-bosh-guid"))
			Expect(instanceGroupGUID).To(Equal("some-resource-guid"))
			Expect(jobConfiguration).To(Equal(api.JobProperties{
				InstanceType: api.InstanceType{
					ID: "some-type",
				},
				FloatingIPs: "1.2.3.4",
			}))
			Expect(logger.PrintfCallCount()).To(Equal(12))
			Expect(logger.PrintfArgsForCall(0)).To(Equal("started configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(1)).To(Equal("finished configuring director options for bosh tile"))
			Expect(logger.PrintfArgsForCall(2)).To(Equal("started configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(3)).To(Equal("finished configuring availability zone options for bosh tile"))
			Expect(logger.PrintfArgsForCall(4)).To(Equal("started configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(5)).To(Equal("finished configuring network options for bosh tile"))
			Expect(logger.PrintfArgsForCall(6)).To(Equal("started configuring network assignment options for bosh tile"))
			Expect(logger.PrintfArgsForCall(7)).To(Equal("finished configuring network assignment options for bosh tile"))
			Expect(logger.PrintfArgsForCall(8)).To(Equal("started configuring resource options for bosh tile"))
			Expect(logger.PrintfArgsForCall(9)).To(Equal("applying resource configuration for the following jobs:"))
			formatStr, formatArg := logger.PrintfArgsForCall(10)
			Expect([]interface{}{formatStr, formatArg}).To(Equal([]interface{}{"\t%s", []interface{}{"resource"}}))
			Expect(logger.PrintfArgsForCall(11)).To(Equal("finished configuring resource options for bosh tile"))
		}

		It("configures the director", func() {
			err := command.Execute([]string{
				"--az-configuration", azConfigurationJSON,
				"--network-assignment", networkAssignmentJSON,
				"--iaas-configuration", iaasConfigurationJSON,
				"--syslog-configuration", syslogConfigurationJSON,
				"--networks-configuration", networksConfigurationJSON,
				"--director-configuration", directorConfigurationJSON,
				"--security-configuration", securityConfigurationJSON,
				"--resource-configuration", resourceConfigurationJSON,
			})
			Expect(err).NotTo(HaveOccurred())

			ExpectDirectorToBeConfiguredCorrectly()
		})

		Context("when the --config flag is set", func() {
			Context("when other flags are set", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--config", "test.yml", "--az-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--networks-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--network-assignment", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--director-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--iaas-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--security-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--syslog-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
					err = command.Execute([]string{"--config", "test.yml", "--resource-configuration", "{}"})
					Expect(err).To(MatchError("config flag can not be passed with another configuration flags"))
				})
			})

			Context("with an invalid config", func() {
				It("configures the director", func() {
					configYAML := `invalidYAML`

					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.WriteString(configYAML)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`could not be parsed as valid configuration: yaml: unmarshal errors`))
				})
			})

			Context("with a valid config", func() {
				It("can interpolate variables into the configuration", func() {
					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.Write(templateConfigurationJSON)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					varsFile, err := ioutil.TempFile("", "vars.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = varsFile.WriteString(`network_name: network`)
					Expect(err).ToNot(HaveOccurred())
					Expect(varsFile.Close()).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--config", configFile.Name(),
						"--vars-file", varsFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					ExpectDirectorToBeConfiguredCorrectly()
				})

				It("returns an error of missing variables", func() {
					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.Write(templateConfigurationJSON)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Expected to find variables"))
				})

				It("configures the director", func() {
					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.Write(completeConfigurationJSON)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					ExpectDirectorToBeConfiguredCorrectly()
				})
			})
		})

		Context("when some director configuration flags are provided", func() {
			It("only updates the config for the provided flags", func() {
				err := command.Execute([]string{
					"--networks-configuration", `{"network": "network-1"}`,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorNetworksCallCount()).To(Equal(1))
				Expect(service.UpdateStagedDirectorNetworksArgsForCall(0)).To(Equal(api.NetworkInput{
					Networks: json.RawMessage(`{"network": "network-1"}`),
				}))
				Expect(service.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(0))
			})
		})

		Context("when no director configuration flags are provided", func() {
			It("does not call the properties function ", func() {
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorNetworksCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(0))
			})
		})

		Context("when an error occurs", func() {
			Context("when flag parser fails", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--foo", "bar"})
					Expect(err).To(MatchError("could not parse configure-director flags: flag provided but not defined: -foo"))
				})
			})

			Context("when configuring availability_zones fails", func() {
				It("returns an error", func() {
					service.UpdateStagedDirectorAvailabilityZonesReturns(errors.New("az endpoint failed"))
					err := command.Execute([]string{"--az-configuration", `{}`})
					Expect(err).To(MatchError("availability zones configuration could not be applied: az endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					service.UpdateStagedDirectorNetworksReturns(errors.New("networks endpoint failed"))
					err := command.Execute([]string{"--networks-configuration", `{}`})
					Expect(err).To(MatchError("networks configuration could not be applied: networks endpoint failed"))
				})
			})

			Context("when configuring networks fails", func() {
				It("returns an error", func() {
					service.UpdateStagedDirectorNetworkAndAZReturns(errors.New("director service failed"))
					err := command.Execute([]string{"--network-assignment", `{}`})
					Expect(err).To(MatchError("network and AZs could not be applied: director service failed"))
				})
			})

			Context("when configuring properties fails", func() {
				It("returns an error", func() {
					service.UpdateStagedDirectorPropertiesReturns(errors.New("properties end point failed"))
					err := command.Execute([]string{"--director-configuration", `{}`})
					Expect(err).To(MatchError("properties could not be applied: properties end point failed"))
				})
			})

			Context("when retrieving staged products fails", func() {
				It("returns an error", func() {
					service.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided top-level resource config is not valid JSON", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{{{`})
					Expect(err).To(MatchError(ContainSubstring("resource-configuration")))
				})
			})

			Context("when retrieving jobs for product fails", func() {
				It("returns an error", func() {
					service.ListStagedProductJobsReturns(nil, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided job does not exist", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{"invalid-resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("invalid-resource")))
				})
			})

			Context("when retrieving existing job config fails", func() {
				It("returns an error", func() {
					service.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{"resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			Context("when user-provided nested resource config is not valid JSON", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--resource-configuration", `{"resource": "%%%"}`})
					Expect(err).To(MatchError(ContainSubstring("resource-configuration")))
				})
			})

			Context("when configuring the job fails", func() {
				It("returns an error", func() {
					service.UpdateStagedProductJobResourceConfigReturns(errors.New("some-error"))
					err := command.Execute([]string{"--resource-configuration", `{"resource": {}}`})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns the usage", func() {
			usage := command.Usage()

			Expect(usage.Description).To(Equal("This authenticated command configures the director."))
			Expect(usage.ShortDescription).To(Equal("configures the director"))
			Expect(usage.Flags).To(Equal(command.Options))
		})
	})
})
