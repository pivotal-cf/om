package commands_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ConfigureDirector", func() {
	var (
		stdout     *gbytes.Buffer
		service    *fakes.ConfigureDirectorService
		command    *commands.ConfigureDirector
		err        error
		config     string
		configFile *os.File
	)

	BeforeEach(func() {
		service = &fakes.ConfigureDirectorService{}
		stdout = gbytes.NewBuffer()
		logger := log.New(stdout, "", 0)
		service.InfoReturns(api.Info{Version: "2.2-build243"}, nil)
		service.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{
				GUID: "p-bosh-guid",
			},
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
		Expect(err).ToNot(HaveOccurred())
		defer configFile.Close()

		_, err = configFile.WriteString(config)
		Expect(err).ToNot(HaveOccurred())
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

			Expect(stdout).To(gbytes.Say("started configuring director options for bosh tile"))
			Expect(stdout).To(gbytes.Say("finished configuring director options for bosh tile"))
			Expect(stdout).To(gbytes.Say("started configuring availability zone options for bosh tile"))
			Expect(stdout).To(gbytes.Say("finished configuring availability zone options for bosh tile"))
			Expect(stdout).To(gbytes.Say("started configuring network options for bosh tile"))
			Expect(stdout).To(gbytes.Say("finished configuring network options for bosh tile"))
			Expect(stdout).To(gbytes.Say("started configuring network assignment options for bosh tile"))
			Expect(stdout).To(gbytes.Say("finished configuring network assignment options for bosh tile"))

			Expect(stdout).To(gbytes.Say("started configuring vm extensions"))
			Expect(stdout).To(gbytes.Say("applying vmextensions configuration for the following:"))
			Expect(stdout).To(gbytes.Say("\ta_vm_extension"))
			Expect(stdout).To(gbytes.Say("\tanother_vm_extension"))

			Expect(stdout).To(gbytes.Say("deleting vm extension some_other_vm_extension"))
			Expect(stdout).To(gbytes.Say("done deleting vm extension some_other_vm_extension"))
			Expect(stdout).To(gbytes.Say("deleting vm extension some_vm_extension"))
			Expect(stdout).To(gbytes.Say("done deleting vm extension some_vm_extension"))
			Expect(stdout).To(gbytes.Say("finished configuring vm extensions"))

			Expect(stdout).To(gbytes.Say("started configuring resource options for bosh tile"))
			Expect(stdout).To(gbytes.Say("finished configuring resource options for bosh tile"))
		}

		It("configures the director", func() {
			err := executeCommand(command, []string{
				"--config", configFile.Name(),
			})
			Expect(err).ToNot(HaveOccurred())

			ExpectDirectorToBeConfiguredCorrectly()
		})

		When("configuring vm types", func() {
			Context("with custom vm types only", func() {
				It("configures the vm types specifically", func() {
					err := executeCommand(command, []string{
						"--config", configFile.Name(),
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(stdout).To(gbytes.Say("creating custom vm types"))
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
					Expect(err).ToNot(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring("if custom_types = true, vm_types must not be empty")))
				})
			})

			Context("setting resource configuration with custom VM types", func() {
				BeforeEach(func() {
					service.ListVMTypesReturns(nil, errors.New("hello"))

				})

				It("does throw an error if the type doesn't exist", func() {
					simpleConfig := `
resource-configuration:
  compilation:
    instance_type:
      id: vmtype5
vmtypes-configuration:
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
					Expect(err).ToNot(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).To(HaveOccurred())
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
					Expect(err).ToNot(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(stdout).To(gbytes.Say("creating custom vm types"))
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
					Expect(err).ToNot(HaveOccurred())
					defer newConfigFile.Close()

					_, err = newConfigFile.WriteString(simpleConfig)
					Expect(err).ToNot(HaveOccurred())

					err = executeCommand(command, []string{
						"--config", newConfigFile.Name(),
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(stdout).To(gbytes.Say("creating custom vm types"))
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

					err = executeCommand(command, []string{
						"--config", configFile.Name(),
					})
					Expect(err).To(MatchError(ContainSubstring(`could not be parsed as valid configuration:`)))
				})
			})

			Context("with a valid config", func() {
				When("the config file(s) contain variables", func() {
					Context("not provided", func() {
						It("returns an error", func() {
							err = executeCommand(command, []string{
								"--config", writeTestConfigFile("vmextensions-configuration: [{name: ((name))}]"),
							})
							Expect(err).To(MatchError(ContainSubstring("Expected to find variables")))
						})
					})

					Context("passed in a file (--vars-file)", func() {
						It("interpolates variables into the configuration", func() {
							err = executeCommand(command, []string{
								"--config", writeTestConfigFile("vmextensions-configuration: [{name: ((name))}]"),
								"--vars-file", writeTestConfigFile("name: network"),
							})
							Expect(err).ToNot(HaveOccurred())

							Expect(service.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
								Name:            "network",
								CloudProperties: json.RawMessage("null"),
							}))
						})
					})

					Context("passed in a var (--var)", func() {
						It("interpolates variables into the configuration", func() {
							err = executeCommand(command, []string{
								"--config", writeTestConfigFile("vmextensions-configuration: [{name: ((name))}]"),
								"--var", "name=network",
							})
							Expect(err).ToNot(HaveOccurred())
							Expect(service.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
								Name:            "network",
								CloudProperties: json.RawMessage("null"),
							}))
						})
					})

					Context("passed as environment variables (--vars-env)", func() {
						It("interpolates variables into the configuration", func() {
							logger := log.New(stdout, "", 0)
							command = commands.NewConfigureDirector(
								func() []string { return []string{"OM_VAR_name=network"} },
								service,
								logger)

							err = executeCommand(command, []string{
								"--config", writeTestConfigFile("vmextensions-configuration: [{name: ((name))}]"),
								"--vars-env", "OM_VAR",
							})
							Expect(err).ToNot(HaveOccurred())
							Expect(service.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
								Name:            "network",
								CloudProperties: json.RawMessage("null"),
							}))
						})

						It("supports the experimental feature of OM_VARS_ENV", func() {
							err := os.Setenv("OM_VARS_ENV", "OM_VAR")
							Expect(err).ToNot(HaveOccurred())
							defer os.Unsetenv("OM_VARS_ENV")

							logger := log.New(stdout, "", 0)

							command = commands.NewConfigureDirector(
								func() []string { return []string{"OM_VAR_name=network"} },
								service,
								logger)

							err = executeCommand(command, []string{
								"--config", writeTestConfigFile("vmextensions-configuration: [{name: ((name))}]"),
							})
							Expect(err).ToNot(HaveOccurred())
							Expect(service.CreateStagedVMExtensionArgsForCall(0)).To(Equal(api.CreateVMExtension{
								Name:            "network",
								CloudProperties: json.RawMessage("null"),
							}))
						})
					})
				})
			})

			Context("with unrecognized top-level-keys", func() {
				It("returns error saying the specified key", func() {
					err = executeCommand(command, []string{
						"--config", writeTestConfigFile(`{"unrecognized-key": {"some-attr": "some-val"}, "unrecognized-other-key": {}, "network-assignment": {"some-attr1": "some-val1"}}`),
					})
					Expect(err).To(MatchError(ContainSubstring(`the config file contains unrecognized keys: "unrecognized-key", "unrecognized-other-key"`)))
				})
			})
		})

		When("no vm_extension configuration is provided", func() {
			It("does not list, create or delete vm extensions", func() {
				err = executeCommand(command, []string{
					"--config", writeTestConfigFile(`{}`),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(service.ListStagedVMExtensionsCallCount()).To(Equal(0))
				Expect(service.CreateStagedVMExtensionCallCount()).To(Equal(0))
				Expect(service.DeleteVMExtensionCallCount()).To(Equal(0))
			})
		})

		When("empty vm_extension configuration is provided", func() {
			It("should delete existing vm extensions", func() {
				err = executeCommand(command, []string{
					"--config", writeTestConfigFile(`vmextensions-configuration: []`),
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(service.ListStagedVMExtensionsCallCount()).To(Equal(1))
				Expect(service.DeleteVMExtensionCallCount()).To(Equal(2))
			})
		})

		When("only some of the configure-director top-level keys are provided", func() {
			It("only updates the config for the provided flags, and sets others to empty", func() {
				err := executeCommand(command, []string{
					"--config", writeTestConfigFile(`{
						"networks-configuration": {
							"network": "network-1"
						}, 
						"properties-configuration": {
							"some-director-assignment": "director"
						}
					}`),
				})
				Expect(err).ToNot(HaveOccurred())

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
			It("returns an error", func() {
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

				err := executeCommand(command, []string{
					"--config", configFile.Name(),
				})

				Expect(err).To(MatchError("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes"))
				Expect(service.ListInstallationsCallCount()).To(Equal(1))
			})
		})

		When("iaas-configurations is set", func() {
			It("configures the director", func() {
				err := executeCommand(command, []string{
					"--config", writeTestConfigFile(`"iaas-configurations": [{"name": "default", "guid": "some-guid"}]`),
					"--ignore-verifier-warnings",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(service.UpdateStagedDirectorIAASConfigurationsCallCount()).To(Equal(1))
				output, ignoreWarnings := service.UpdateStagedDirectorIAASConfigurationsArgsForCall(0)
				Expect(string(output)).To(MatchYAML(`[{guid: some-guid, name: default}]`))
				Expect(ignoreWarnings).To(BeTrue())
			})

			When("setting iaas configurations fails", func() {
				It("returns an error", func() {
					service.UpdateStagedDirectorIAASConfigurationsReturns(errors.New("iaas failed"))
					err := executeCommand(command, []string{"--config", writeTestConfigFile(`"iaas-configurations": [{"name": "default", "guid": "some-guid"}]`)})
					Expect(err).To(MatchError("iaas configurations could not be completed: iaas failed"))
				})
			})

			When("iaas-configurations is used with a version of ops manager below 2.6", func() {
				It("returns an error", func() {
					versions := []string{"2.1-build.326", "1.12-build99"}
					for _, version := range versions {
						service.InfoReturns(api.Info{Version: version}, nil)

						err := executeCommand(command, []string{"--config", writeTestConfigFile(`"iaas-configurations": [{"name": "default", "guid": "some-guid"}]`)})
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("\"iaas-configurations\" is only available with Ops Manager 2.2 or later: you are running %s", version)))
					}
				})
			})
		})

		When("iaas-configurations and properties-configuration.iaas-configuration are both set", func() {
			It("returns an error", func() {
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"iaas-configurations": [], "properties-configuration": {"iaas-configuration": {}}}`)})
				Expect(err).To(MatchError("iaas-configurations cannot be used with properties-configuration.iaas-configurations\nPlease only use one implementation."))
			})
		})

		When("configuring availability_zones fails", func() {
			It("returns an error", func() {
				service.UpdateStagedDirectorAvailabilityZonesReturns(errors.New("az endpoint failed"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"az-configuration": {}}`)})
				Expect(err).To(MatchError("availability zones configuration could not be applied: az endpoint failed"))
			})
		})

		When("configuring networks fails", func() {
			It("returns an error", func() {
				service.UpdateStagedDirectorNetworksReturns(errors.New("networks endpoint failed"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"networks-configuration": {}}`)})
				Expect(err).To(MatchError("networks configuration could not be applied: networks endpoint failed"))
			})
		})

		When("configuring networks fails", func() {
			It("returns an error", func() {
				service.UpdateStagedDirectorNetworkAndAZReturns(errors.New("director service failed"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"network-assignment": {}}`)})
				Expect(err).To(MatchError("network and AZs could not be applied: director service failed"))
			})
		})

		When("configuring properties fails", func() {
			It("returns an error", func() {
				service.UpdateStagedDirectorPropertiesReturns(errors.New("properties end point failed"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"properties-configuration": {"director_configuration": {}}}`)})
				Expect(err).To(MatchError("properties could not be applied: properties end point failed"))
			})
		})

		When("retrieving staged products fails", func() {
			It("returns an error", func() {
				service.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"resource-configuration": {}}`)})
				Expect(err).To(MatchError(ContainSubstring("some-error")))
			})
		})

		When("user-provided top-level resource config is not valid JSON", func() {
			It("returns an error", func() {
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"resource-configuration": {{{{}`)})
				Expect(err).To(MatchError(ContainSubstring("did not find expected ',' or '}'")))
			})
		})

		When("configuring the job fails", func() {
			It("returns an error", func() {
				service.ConfigureJobResourceConfigReturns(errors.New("some-error"))
				err := executeCommand(command, []string{"--config", writeTestConfigFile(`{"resource-configuration": {"resource": {}}}`)})
				Expect(err).To(MatchError(ContainSubstring("some-error")))
			})
		})
	})
})
