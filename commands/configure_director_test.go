package commands_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigureDirector", func() {
	var (
		logger     *fakes.Logger
		service    *fakes.ConfigureDirectorService
		command    commands.ConfigureDirector
		err        error
		config     string
		configFile *os.File
	)

	BeforeEach(func() {
		service = &fakes.ConfigureDirectorService{}
		logger = &fakes.Logger{}
		service.InfoReturns(api.Info{Version: "2.2-build243"}, nil)
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
		service.ListStagedVMExtensionsReturns([]api.VMExtension{
			{Name: "some_vm_extension"},
			{Name: "some_other_vm_extension"},
		}, nil)

		service.ListInstallationsReturns([]api.InstallationsServiceOutput{
			{
				ID:         999,
				Status:     "succeeded",
				Logs:       "",
				StartedAt:  nil,
				FinishedAt: nil,
				UserName:   "admin",
			},
		}, nil)

		command = commands.NewConfigureDirector(
			func() []string { return []string{} },
			service,
			logger)
	})

	JustBeforeEach(func() {
		configFile, err = ioutil.TempFile("", "config.yml")
		Expect(err).NotTo(HaveOccurred())
		defer configFile.Close()

		_, err = configFile.WriteString(config)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			config = `
---
network-assignment:
  network:
    name: network
  singleton_availability_zone:
    name: singleton
az-configuration:
- clusters:
  - cluster: pizza-boxes
  name: AZ1
networks-configuration:
  network: network-1
resource-configuration:
  resource:
    instance_type:
      id: some-type
vmextensions-configuration:
- name: a_vm_extension
  cloud_properties:
    source_dest_check: false
- name: another_vm_extension
  cloud_properties:
    foo: bar
properties-configuration:
  dns_configuration:
    recurse: "true"
  syslog_configuration:
    some-syslog-assignment: syslog
  security_configuration:
    some-security-assignment: security
  iaas_configuration:
    some-iaas-assignment: iaas
  director_configuration:
    some-director-assignment: director
vmtypes-configuration:
  custom_only: true
  vm_types:
  - name: vmtype3
    cpu: 1
    ram: 2048
    ephemeral_disk: 10240
  - name: vmtype4
    cpu: 2
    ram: 4096
    ephemeral_disk: 20480
`
		})

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
			Expect(string(service.UpdateStagedDirectorPropertiesArgsForCall(0))).To(MatchJSON(
				`{
					"director_configuration":{"some-director-assignment":"director"},
					"iaas_configuration":{"some-iaas-assignment":"iaas"},
					"security_configuration":{"some-security-assignment":"security"},
					"syslog_configuration":{"some-syslog-assignment":"syslog"},
					"dns_configuration": {"recurse":"true"}
				}`,
			))
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

			Expect(service.ListStagedVMExtensionsCallCount()).To(Equal(1))
			Expect(service.CreateStagedVMExtensionCallCount()).To(Equal(2))
			Expect(service.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
				Name:            "a_vm_extension",
				CloudProperties: json.RawMessage(`{"source_dest_check":false}`),
			}))
			Expect(service.CreateStagedVMExtensionArgsForCall(1)).To(Equal(api.CreateVMExtension{
				Name:            "another_vm_extension",
				CloudProperties: json.RawMessage(`{"foo":"bar"}`),
			}))

			Expect(service.DeleteVMExtensionCallCount()).To(Equal(2))
			deletedExtensions := []string{service.DeleteVMExtensionArgsForCall(0), service.DeleteVMExtensionArgsForCall(1)}
			Expect(deletedExtensions).To(ContainElement("some_other_vm_extension"))
			Expect(deletedExtensions).To(ContainElement("some_vm_extension"))

			Expect(logger.PrintfCallCount()).To(BeNumerically(">=", 21))
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
			Expect(logger.PrintfArgsForCall(12)).To(Equal("started configuring vm extensions"))
			Expect(logger.PrintfArgsForCall(13)).To(Equal("applying vmextensions configuration for the following:"))
			formatStr, formatArg = logger.PrintfArgsForCall(14)
			Expect([]interface{}{formatStr, formatArg}).To(Equal([]interface{}{"\t%s", []interface{}{"a_vm_extension"}}))
			formatStr, formatArg = logger.PrintfArgsForCall(15)
			Expect([]interface{}{formatStr, formatArg}).To(Equal([]interface{}{"\t%s", []interface{}{"another_vm_extension"}}))

			expectedLogs := make(map[interface{}][]string)
			formatStr1, _ := logger.PrintfArgsForCall(16)
			formatStr2, formatArg := logger.PrintfArgsForCall(17)
			expectedLogs[formatArg[0]] = []string{formatStr1, formatStr2}
			formatStr1, _ = logger.PrintfArgsForCall(18)
			formatStr2, formatArg = logger.PrintfArgsForCall(19)
			expectedLogs[formatArg[0]] = []string{formatStr1, formatStr2}
			Expect(expectedLogs).To(HaveKey("some_other_vm_extension"))
			Expect(expectedLogs).To(HaveKey("some_vm_extension"))
			Expect(expectedLogs["some_vm_extension"]).To(ContainElement("deleting vm extension %s"))
			Expect(expectedLogs["some_vm_extension"]).To(ContainElement("done deleting vm extension %s"))
			Expect(expectedLogs["some_other_vm_extension"]).To(ContainElement("deleting vm extension %s"))
			Expect(expectedLogs["some_other_vm_extension"]).To(ContainElement("done deleting vm extension %s"))

			Expect(logger.PrintfArgsForCall(20)).To(Equal("finished configuring vm extensions"))
		}

		It("configures the director", func() {
			err := command.Execute([]string{
				"--config", configFile.Name(),
			})
			Expect(err).NotTo(HaveOccurred())

			ExpectDirectorToBeConfiguredCorrectly()
		})

		When("configuring vm types", func() {
			Context("with custom vm types only", func() {
				It("configures the vm types specifically", func() {
					err := command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCallCount()).To(Equal(22))
					Expect(logger.PrintfArgsForCall(21)).To(Equal("creating custom vm types"))
					Expect(service.ListVMTypesCallCount()).To(Equal(0))
					Expect(service.DeleteCustomVMTypesCallCount()).To(Equal(0))
					Expect(service.CreateCustomVMTypesCallCount()).To(Equal(1))
					Expect(service.CreateCustomVMTypesArgsForCall(0).VMTypes).To(HaveLen(2))
				})

				It("errors if there are no vm types specified", func() {
					simpleConfig := `---
vmtypes-configuration:
  custom_only: true
`
					newConfigFile, err := ioutil.TempFile("", "config.yml")
					Expect(err).NotTo(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).NotTo(HaveOccurred())

					err = command.Execute([]string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("if custom_types = true, vm_types must not be empty"))
				})
			})

			Context("with custom and builtin VM types", func() {
				BeforeEach(func() {
					service.ListVMTypesReturns([]api.VMType{
						{CreateVMType: api.CreateVMType{Name: "vmtype1", CPU: 2, RAM: 4096}, BuiltIn: true},
						{CreateVMType: api.CreateVMType{Name: "vmtype2", CPU: 2, RAM: 8192}, BuiltIn: true},
					}, nil)
				})

				It("adds custom vm types to existing types", func() {
					simpleConfig := `vmtypes-configuration:
  vm_types:
  - name: vmtype3
    cpu: 1
    ram: 2048
    ephemeral_disk: 10240
  - name: vmtype4
    cpu: 2
    ram: 4096
    ephemeral_disk: 20480
`
					newConfigFile, err := ioutil.TempFile("", "config.yml")
					Expect(err).NotTo(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).NotTo(HaveOccurred())

					err = command.Execute([]string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCallCount()).To(Equal(1))
					Expect(logger.PrintfArgsForCall(0)).To(Equal("creating custom vm types"))
					Expect(service.ListVMTypesCallCount()).To(Equal(1))
					Expect(service.DeleteCustomVMTypesCallCount()).To(Equal(1))
					Expect(service.CreateCustomVMTypesCallCount()).To(Equal(1))
					Expect(service.CreateCustomVMTypesArgsForCall(0).VMTypes).To(HaveLen(4))
				})

				It("overwrites existing vm types", func() {
					simpleConfig := `vmtypes-configuration:
  vm_types:
  - name: vmtype2
    cpu: 1
    ram: 2048
    ephemeral_disk: 10240
  - name: vmtype3
    cpu: 2
    ram: 4096
    ephemeral_disk: 20480
`
					newConfigFile, err := ioutil.TempFile("", "config.yml")
					Expect(err).NotTo(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).NotTo(HaveOccurred())

					err = command.Execute([]string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCallCount()).To(Equal(1))
					Expect(logger.PrintfArgsForCall(0)).To(Equal("creating custom vm types"))
					Expect(service.ListVMTypesCallCount()).To(Equal(1))
					Expect(service.DeleteCustomVMTypesCallCount()).To(Equal(1))
					Expect(service.CreateCustomVMTypesCallCount()).To(Equal(1))
					Expect(service.CreateCustomVMTypesArgsForCall(0).VMTypes).To(HaveLen(3))
					Expect(service.CreateCustomVMTypesArgsForCall(0).VMTypes[1].CPU).To(BeEquivalentTo(1))
				})
			})
		})

		When("the --config flag is set", func() {
			Context("with an invalid config", func() {
				It("does not configure the director", func() {
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
					Expect(err.Error()).To(ContainSubstring(`could not be parsed as valid configuration:`))
				})
			})

			Context("with a valid config", func() {
				var (
					completeConfigurationJSON []byte
					templateConfigurationJSON []byte
				)

				BeforeEach(func() {
					azConfiguration := []interface{}{map[string]interface{}{
						"name": "AZ1",
						"clusters": []interface{}{map[string]interface{}{
							"cluster": "pizza-boxes",
						}},
					}}
					iaasConfiguration := map[string]interface{}{
						"some-iaas-assignment": "iaas",
					}
					networkAssignment := map[string]interface{}{
						"network": map[string]interface{}{
							"name": "network",
						},
						"singleton_availability_zone": map[string]interface{}{
							"name": "singleton",
						},
					}
					syslogConfiguration := map[string]interface{}{
						"some-syslog-assignment": "syslog",
					}
					networksConfiguration := map[string]interface{}{
						"network": "network-1",
					}
					directorConfiguration := map[string]interface{}{
						"some-director-assignment": "director",
					}
					securityConfiguration := map[string]interface{}{
						"some-security-assignment": "security",
					}
					resourceConfiguration := map[string]interface{}{
						"resource": map[string]interface{}{
							"instance_type": map[string]interface{}{
								"id": "some-type",
							},
						},
					}
					vmextensionConfig := []map[string]interface{}{
						{
							"name": "a_vm_extension",
							"cloud_properties": map[string]interface{}{
								"source_dest_check": false,
							},
						},

						{
							"name": "another_vm_extension",
							"cloud_properties": map[string]interface{}{
								"foo": "bar",
							},
						},
					}

					templateNetworkAssign := map[string]interface{}{
						"network": map[string]interface{}{
							"name": "((network_name))",
						},
						"singleton_availability_zone": map[string]interface{}{
							"name": "singleton",
						},
					}

					dnsConfig := map[string]interface{}{
						"recurse": "true",
					}

					configurationMAP := map[string]interface{}{}
					configurationMAP["network-assignment"] = networkAssignment
					configurationMAP["az-configuration"] = azConfiguration
					configurationMAP["networks-configuration"] = networksConfiguration
					configurationMAP["resource-configuration"] = resourceConfiguration
					configurationMAP["vmextensions-configuration"] = vmextensionConfig

					configurationMAP["properties-configuration"] = map[string]interface{}{
						"director_configuration": directorConfiguration,
						"iaas_configuration":     iaasConfiguration,
						"security_configuration": securityConfiguration,
						"dns_configuration":      dnsConfig,
						"syslog_configuration":   syslogConfiguration,
					}

					completeConfigurationJSON, err = json.Marshal(configurationMAP)
					Expect(err).NotTo(HaveOccurred())

					configurationMAP["network-assignment"] = templateNetworkAssign
					templateConfigurationJSON, err = json.Marshal(configurationMAP)
					Expect(err).NotTo(HaveOccurred())
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

				When("the config file(s) contain variables", func() {

					Context("not provided", func() {
						It("returns an error", func() {
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
					})

					Context("passed in a file (--vars-file)", func() {
						It("interpolates variables into the configuration", func() {
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
					})

					Context("passed in a var (--var)", func() {
						It("interpolates variables into the configuration", func() {
							configFile, err := ioutil.TempFile("", "config.yaml")
							Expect(err).ToNot(HaveOccurred())
							_, err = configFile.Write(templateConfigurationJSON)
							Expect(err).ToNot(HaveOccurred())
							Expect(configFile.Close()).ToNot(HaveOccurred())

							err = command.Execute([]string{
								"--config", configFile.Name(),
								"--var", "network_name=network",
							})
							Expect(err).NotTo(HaveOccurred())

							ExpectDirectorToBeConfiguredCorrectly()
						})
					})

					Context("passed as environment variables (--vars-env)", func() {
						It("interpolates variables into the configuration", func() {

							command = commands.NewConfigureDirector(
								func() []string { return []string{"OM_VAR_network_name=network"} },
								service,
								logger)

							configFile, err := ioutil.TempFile("", "config.yaml")
							Expect(err).ToNot(HaveOccurred())
							_, err = configFile.Write(templateConfigurationJSON)
							Expect(err).ToNot(HaveOccurred())
							Expect(configFile.Close()).ToNot(HaveOccurred())

							err = command.Execute([]string{
								"--config", configFile.Name(),
								"--vars-env", "OM_VAR",
							})
							Expect(err).NotTo(HaveOccurred())

							ExpectDirectorToBeConfiguredCorrectly()
						})
					})

				})

			})

			Context("with unrecognized top-level-keys", func() {
				It("returns error saying the specified key", func() {
					configYAML := `{"unrecognized-key": {"some-attr": "some-val"}, "unrecognized-other-key": {}, "network-assignment": {"some-attr1": "some-val1"}}`
					configFile, err := ioutil.TempFile("", "config.yaml")
					Expect(err).ToNot(HaveOccurred())
					_, err = configFile.WriteString(configYAML)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFile.Close()).ToNot(HaveOccurred())

					err = command.Execute([]string{
						"--config", configFile.Name(),
					})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`the config file contains unrecognized keys: "unrecognized-key", "unrecognized-other-key"`))
				})
			})
		})

		When("no vm_extension configuration is provided", func() {
			It("does not list, create or delete vm extensions", func() {
				configurationMAP := map[string]interface{}{}

				completeConfigurationJSON, err := json.Marshal(configurationMAP)
				Expect(err).NotTo(HaveOccurred())
				configFile, err := ioutil.TempFile("", "config.yaml")
				Expect(err).ToNot(HaveOccurred())
				_, err = configFile.Write(completeConfigurationJSON)
				Expect(err).ToNot(HaveOccurred())
				Expect(configFile.Close()).ToNot(HaveOccurred())

				err = command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.ListStagedVMExtensionsCallCount()).To(Equal(0))
				Expect(service.CreateStagedVMExtensionCallCount()).To(Equal(0))
				Expect(service.DeleteVMExtensionCallCount()).To(Equal(0))
			})
		})

		When("empty vm_extension configuration is provided", func() {
			It("should delete existing vm extensions", func() {
				configFile, err := ioutil.TempFile("", "config.yaml")
				Expect(err).ToNot(HaveOccurred())
				_, err = configFile.Write([]byte(`vmextensions-configuration: []`))
				Expect(err).ToNot(HaveOccurred())
				Expect(configFile.Close()).ToNot(HaveOccurred())

				err = command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.ListStagedVMExtensionsCallCount()).To(Equal(1))
				Expect(service.DeleteVMExtensionCallCount()).To(Equal(2))
			})
		})

		When("only some of the configure-director top-level keys are provided", func() {
			BeforeEach(func() {
				config = `{"networks-configuration":{"network":"network-1"},"properties-configuration":{"some-director-assignment":"director"}}`
			})

			It("only updates the config for the provided flags, and sets others to empty", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(service.UpdateStagedDirectorAvailabilityZonesCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorNetworksCallCount()).To(Equal(1))
				Expect(service.UpdateStagedDirectorNetworksArgsForCall(0)).To(Equal(api.NetworkInput{
					Networks: json.RawMessage(`{"network":"network-1"}`),
				}))
				Expect(service.UpdateStagedDirectorNetworkAndAZCallCount()).To(Equal(0))
				Expect(service.UpdateStagedDirectorPropertiesCallCount()).To(Equal(1))
				Expect(service.UpdateStagedDirectorPropertiesArgsForCall(0)).To(Equal(api.DirectorProperties(
					`{"some-director-assignment":"director"}`,
				)))
			})
		})

		When("there is a running installation", func() {
			BeforeEach(func() {
				service.ListInstallationsReturns([]api.InstallationsServiceOutput{
					{
						ID:         999,
						Status:     "running",
						Logs:       "",
						StartedAt:  nil,
						FinishedAt: nil,
						UserName:   "admin",
					},
				}, nil)
			})
			It("returns an error", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).To(MatchError("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes"))
				Expect(service.ListInstallationsCallCount()).To(Equal(1))
			})
		})

		When("an error occurs", func() {
			When("no director configuration flags are provided", func() {
				It("returns an error ", func() {
					err := command.Execute([]string{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("missing required flag \"--config\""))
				})
			})

			When("flag parser fails", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--foo", "bar"})
					Expect(err).To(MatchError("could not parse configure-director flags: flag provided but not defined: -foo"))
				})
			})

			When("configuring availability_zones fails", func() {
				BeforeEach(func() {
					config = `{"az-configuration": {}}`
				})

				It("returns an error", func() {
					service.UpdateStagedDirectorAvailabilityZonesReturns(errors.New("az endpoint failed"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("availability zones configuration could not be applied: az endpoint failed"))
				})
			})

			When("configuring networks fails", func() {
				BeforeEach(func() {
					config = `{"networks-configuration": {}}`
				})

				It("returns an error", func() {
					service.UpdateStagedDirectorNetworksReturns(errors.New("networks endpoint failed"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("networks configuration could not be applied: networks endpoint failed"))
				})
			})

			When("configuring networks fails", func() {
				BeforeEach(func() {
					config = `{"network-assignment": {}}`
				})

				It("returns an error", func() {
					service.UpdateStagedDirectorNetworkAndAZReturns(errors.New("director service failed"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("network and AZs could not be applied: director service failed"))
				})
			})

			When("configuring properties fails", func() {
				BeforeEach(func() {
					config = `{"properties-configuration": {"director_configuration": {}}}`
				})

				It("returns an error", func() {
					service.UpdateStagedDirectorPropertiesReturns(errors.New("properties end point failed"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("properties could not be applied: properties end point failed"))
				})
			})

			When("retrieving staged products fails", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {}}`
				})

				It("returns an error", func() {
					service.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			When("user-provided top-level resource config is not valid JSON", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {{{{}`
				})

				It("returns an error", func() {
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("did not find expected ',' or '}'")))
				})
			})

			When("retrieving jobs for product fails", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {}}`
				})

				It("returns an error", func() {
					service.ListStagedProductJobsReturns(nil, errors.New("some-error"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			When("user-provided job does not exist", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {"invalid-resource": {}}}`
				})

				It("returns an error", func() {
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("invalid-resource")))
				})
			})

			When("retrieving existing job config fails", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {"resource": {}}}`
				})

				It("returns an error", func() {
					service.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})

			When("user-provided nested resource config is not valid JSON", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {"resource": "%%%"}}`
				})

				It("returns an error", func() {
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("could not decode resource-configuration json for job 'resource'")))
				})
			})

			When("configuring the job fails", func() {
				BeforeEach(func() {
					config = `{"resource-configuration": {"resource": {}}}`
				})

				It("returns an error", func() {
					service.UpdateStagedProductJobResourceConfigReturns(errors.New("some-error"))
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError(ContainSubstring("some-error")))
				})
			})
		})

		When("iaas-configurations is set", func() {
			BeforeEach(func() {
				config = `
iaas-configurations:
- {
	  guid: some-guid,
	  name: default,
  }
`
			})

			It("configures the director", func() {
				err := command.Execute([]string{
					"--config", configFile.Name(),
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.UpdateStagedDirectorIAASConfigurationsCallCount()).To(Equal(1))
				Expect(string(service.UpdateStagedDirectorIAASConfigurationsArgsForCall(0))).To(MatchYAML(`[{guid: some-guid, name: default}]`))
			})

			Context("failure cases", func() {
				When("setting iaas configurations fails", func() {
					It("returns an error", func() {
						service.UpdateStagedDirectorIAASConfigurationsReturns(errors.New("iaas failed"))
						err := command.Execute([]string{"--config", configFile.Name()})
						Expect(err).To(MatchError("iaas configurations could not be completed: iaas failed"))
					})
				})

				When("iaas-configurations is used with a version of ops manager below 2.6", func() {
					It("returns an error", func() {
						versions := []string{"2.1-build.326", "1.12-build99"}
						for _, version := range versions {
							service.InfoReturns(api.Info{Version: version}, nil)

							err := command.Execute([]string{"--config", configFile.Name()})
							Expect(err).To(MatchError(fmt.Sprintf("\"iaas-configurations\" is only available with Ops Manager 2.2 or later: you are running %s", version)))
						}
					})
				})
			})

			When("iaas-configurations and properties-configuration.iaas-configuration are both set", func() {
				BeforeEach(func() {
					config = `{"iaas-configurations": [], "properties-configuration": {"iaas-configuration": {}}}`
				})

				It("returns an error", func() {
					err := command.Execute([]string{"--config", configFile.Name()})
					Expect(err).To(MatchError("iaas-configurations cannot be used with properties-configuration.iaas-configurations\nPlease only use one implementation."))
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
