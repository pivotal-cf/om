package commands_test

import (
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("StagedDirectorConfig", func() {
	var (
		stdout      *fakes.Logger
		stderr      *fakes.Logger
		fakeService *fakes.StagedDirectorConfigService
	)

	BeforeEach(func() {
		stdout = &fakes.Logger{}
		stderr = &fakes.Logger{}
		fakeService = &fakes.StagedDirectorConfigService{}
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			expectedDirectorAZs := api.AvailabilityZonesOutput{
				AvailabilityZones: []api.AvailabilityZoneOutput{
					{
						Name:                  "some-az",
						IAASConfigurationName: "some-iaas",
					}, {
						Name:                  "some-other-az",
						IAASConfigurationName: "some-other-iaas",
					},
				},
			}
			fakeService.GetStagedDirectorAvailabilityZonesReturns(expectedDirectorAZs, nil)

			expectedDirectorProperties := map[string]interface{}{
				"director_configuration": map[string]interface{}{
					"filtered_key": "filtered_key",
					"max_threads":  5,
					"encryption": map[string]interface{}{
						"providers": map[string]interface{}{
							"partition_password": "some_password",
							"client_certificate": "user_provided_cert",
							"client_key":         "user_provided_key",
							"client_user":        "user",
						},
					},
				},
				"iaas_configuration": map[interface{}]interface{}{
					"project": "project-id",
					"key":     "some-key",
				},
				"syslog_configuration": map[string]interface{}{
					"syslogconfig": "awesome",
				},
				"security_configuration": map[string]interface{}{
					"trusted_certificates": "some-certificate",
				},
			}
			fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties, nil)

			expectedNetworks := api.NetworksConfigurationOutput{
				Networks: []api.NetworkConfigurationOutput{
					{
						Name: "network-1",
					},
				},
			}
			fakeService.GetStagedDirectorNetworksReturns(expectedNetworks, nil)

			fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
				Product: api.StagedProduct{
					GUID: "p-bosh-guid",
					Type: "director",
				},
			}, nil)

			fakeService.GetStagedProductNetworksAndAZsReturns(map[string]interface{}{
				"network": map[string]interface{}{
					"name": "network-1",
				},
				"singleton_availability_zone": map[string]interface{}{
					"name": "some-az",
				},
			}, nil)

			fakeService.ListStagedProductJobsReturns(map[string]string{
				"some-job": "some-job-guid",
			}, nil)

			fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{
				"instances": 1.0,
				"instance_type": map[string]interface{}{
					"id": "automatic",
				},
				"additional_vm_extensions": []string{"some-vm-extension"},
			}, nil)

			expectedVMExtensions := []api.VMExtension{
				{
					Name: "vm_ext1",
					CloudProperties: map[string]interface{}{
						"source_dest_check": false,
					},
				},
				{
					Name: "vm_ext2",
					CloudProperties: map[string]interface{}{
						"key_name": "operations_keypair",
					},
				},
			}
			fakeService.ListStagedVMExtensionsReturns(expectedVMExtensions, nil)

			fakeService.GetStagedDirectorIaasConfigurationsReturns(nil, nil)

			fakeService.ListVMTypesReturns([]api.VMType{
				{CreateVMType: api.CreateVMType{Name: "vm-type1", CPU: 1, RAM: 2048, EphemeralDisk: 10240}, BuiltIn: true},
				{CreateVMType: api.CreateVMType{Name: "vm-type2", CPU: 2, RAM: 2048, EphemeralDisk: 10240}, BuiltIn: true},
			}, nil)
		})

		It("writes a complete config file with filtered sensitive fields to stdout", func() {
			command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			output := stdout.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  syslog_configuration:
    syslogconfig: awesome
  security_configuration:
    trusted_certificates: some-certificate
  director_configuration:
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
vmtypes-configuration: {}
`)))
		})

		It("includes custom vm types when present", func() {
			fakeService.ListVMTypesReturns([]api.VMType{
				{CreateVMType: api.CreateVMType{Name: "vm-type1", CPU: 1, RAM: 2048, EphemeralDisk: 10240}, BuiltIn: false},
				{CreateVMType: api.CreateVMType{Name: "vm-type2", CPU: 2, RAM: 2048, EphemeralDisk: 10240}, BuiltIn: false},
			}, nil)

			command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
			err := command.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			output := stdout.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  syslog_configuration:
    syslogconfig: awesome
  security_configuration:
    trusted_certificates: some-certificate
  director_configuration:
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
vmtypes-configuration:
  custom_only: true
  vm_types:
  - name: vm-type1
    cpu: 1
    ram: 2048
    ephemeral_disk: 10240
  - name: vm-type2
    cpu: 2
    ram: 2048
    ephemeral_disk: 10240
`)))
		})

		It("doesn't redact values when --no-redact is passed", func() {
			command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
			err := command.Execute([]string{"--no-redact"})
			Expect(err).ToNot(HaveOccurred())

			invocations := fakeService.Invocations()["GetStagedDirectorProperties"]
			Expect(invocations[0]).To(Equal([]interface{}{false}))
		})

		When("getting availability_zones returns an empty array", func() {
			It("doesn't return the az in the config", func() {
				fakeService.GetStagedDirectorAvailabilityZonesReturns(api.AvailabilityZonesOutput{}, nil)
				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
vmtypes-configuration: {}
`)))
			})
		})

		It("does not include guid in iaas-configurations", func() {
			fakeService.GetStagedDirectorIaasConfigurationsReturns(map[string][]map[string]interface{}{
				"iaas_configurations": {
					{
						"guid": "some-guid",
						"name": "default",
					},
					{
						"guid": "some-other-guid",
						"name": "configTwo",
					},
				},
			}, nil)

			command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
			err := command.Execute([]string{
				"--no-redact",
			})
			Expect(err).ToNot(HaveOccurred())

			output := stdout.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
iaas-configurations:
  - name: default
  - name: configTwo
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: filtered_key
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
        partition_password: some_password
        client_key: user_provided_key
        client_user: user
vmtypes-configuration: {}
`)))
		})

		It("does not include guid in properties-configuration.iaas_configuration", func() {
			expectedDirectorProperties := map[string]interface{}{
				"director_configuration": map[string]interface{}{
					"filtered_key": "filtered_key",
					"max_threads":  5,
					"encryption": map[string]interface{}{
						"providers": map[string]interface{}{
							"partition_password": "some_password",
							"client_certificate": "user_provided_cert",
							"client_key":         "user_provided_key",
							"client_user":        "user",
						},
					},
				},
				"iaas_configuration": map[interface{}]interface{}{
					"guid": "some-guid",
					"key":  "some-key",
				},
				"syslog_configuration": map[string]interface{}{
					"syslogconfig": "awesome",
				},
				"security_configuration": map[string]interface{}{
					"trusted_certificates": "some-certificate",
				},
			}
			fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties, nil)

			command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
			err := command.Execute([]string{
				"--no-redact",
			})
			Expect(err).ToNot(HaveOccurred())

			output := stdout.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  iaas_configuration:
    key: some-key
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: filtered_key
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
        partition_password: some_password
        client_key: user_provided_key
        client_user: user
vmtypes-configuration: {}
`)))
		})

		Describe("with --no-redact", func() {
			It("Includes the filtered fields when printing to stdout", func() {
				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--no-redact",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: filtered_key
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
        partition_password: some_password
        client_key: user_provided_key
        client_user: user
  iaas_configuration:
    key: some-key
    project: project-id  
vmtypes-configuration: {}
`)))
			})

			It("fetches iaas_configurations from /iaas_configurations endpoint if multi-iaas is supported", func() {
				fakeService.GetStagedDirectorIaasConfigurationsReturns(map[string][]map[string]interface{}{
					"iaas_configurations": {
						{
							"guid":                       "some-guid",
							"name":                       "default",
							"project":                    "my-google-project",
							"associated_service_account": "my-google-service-account",
							"auth_json":                  "****",
						},
					},
				}, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--no-redact",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
iaas-configurations:
  - name: default
    project: my-google-project
    associated_service_account: my-google-service-account
    auth_json: "****"
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: filtered_key
    max_threads: 5
    encryption:
      providers:
        client_certificate: user_provided_cert
        partition_password: some_password
        client_key: user_provided_key
        client_user: user
vmtypes-configuration: {}
`)))
			})
		})

		Describe("with --include-placeholders", func() {
			It("Includes the placeholder fields when printing to stdout", func() {
				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: ((properties-configuration_director_configuration_filtered_key))
    encryption:
      providers:
        client_certificate: user_provided_cert
        client_key: ((properties-configuration_director_configuration_encryption_providers_client_key))
        client_user: ((properties-configuration_director_configuration_encryption_providers_client_user))
        partition_password: ((properties-configuration_director_configuration_encryption_providers_partition_password))
    max_threads: 5
  iaas_configuration:
    project: ((properties-configuration_iaas_configuration_project))
    key: ((properties-configuration_iaas_configuration_key))  
vmtypes-configuration: {}
`)))
			})

			It("with multiple iaas's configured, includes placeholders", func() {
				fakeService.GetStagedDirectorIaasConfigurationsReturns(map[string][]map[string]interface{}{
					"iaas_configurations": {
						{
							"guid":                       "some-guid",
							"name":                       "default",
							"project":                    "my-google-project",
							"associated_service_account": "my-google-service-account",
							"auth_json":                  "****",
						}, {
							"guid":                       "another-guid",
							"name":                       "meow",
							"project":                    "my-other-google-project",
							"associated_service_account": "my-other-google-service-account",
							"auth_json":                  "****",
						},
					},
				}, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: ((properties-configuration_director_configuration_filtered_key))
    encryption:
      providers:
        client_certificate: user_provided_cert
        client_key: ((properties-configuration_director_configuration_encryption_providers_client_key))
        client_user: ((properties-configuration_director_configuration_encryption_providers_client_user))
        partition_password: ((properties-configuration_director_configuration_encryption_providers_partition_password))
    max_threads: 5
iaas-configurations:
  - name: ((iaas-configurations_0_name))
    project: ((iaas-configurations_0_project))
    associated_service_account: ((iaas-configurations_0_associated_service_account))
    auth_json: ((iaas-configurations_0_auth_json))
  - name: ((iaas-configurations_1_name))
    project: ((iaas-configurations_1_project))
    associated_service_account: ((iaas-configurations_1_associated_service_account))
    auth_json: ((iaas-configurations_1_auth_json))
vmtypes-configuration: {}
`)))
			})

			It("is able to handle a bool", func() {
				fakeService.GetStagedDirectorIaasConfigurationsReturns(map[string][]map[string]interface{}{
					"iaas_configurations": {
						{
							"bosh-truthy": true,
						},
					},
				}, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: ((properties-configuration_director_configuration_filtered_key))
    encryption:
      providers:
        client_certificate: user_provided_cert
        client_key: ((properties-configuration_director_configuration_encryption_providers_client_key))
        client_user: ((properties-configuration_director_configuration_encryption_providers_client_user))
        partition_password: ((properties-configuration_director_configuration_encryption_providers_partition_password))
    max_threads: 5
iaas-configurations:
  - bosh-truthy: ((iaas-configurations_0_bosh-truthy))
vmtypes-configuration: {}
`)))
			})

			It("is able to handle an int", func() {
				fakeService.GetStagedDirectorIaasConfigurationsReturns(map[string][]map[string]interface{}{
					"iaas_configurations": {
						{
							"iaas-int": 1,
						},
					},
				}, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  security_configuration:
    trusted_certificates: some-certificate
  syslog_configuration:
    syslogconfig: awesome
  director_configuration:
    filtered_key: ((properties-configuration_director_configuration_filtered_key))
    encryption:
      providers:
        client_certificate: user_provided_cert
        client_key: ((properties-configuration_director_configuration_encryption_providers_client_key))
        client_user: ((properties-configuration_director_configuration_encryption_providers_client_user))
        partition_password: ((properties-configuration_director_configuration_encryption_providers_partition_password))
    max_threads: 5
iaas-configurations:
  - iaas-int: ((iaas-configurations_0_iaas-int))
vmtypes-configuration: {}
`)))
			})

			It("is able to handle an nested map[interface{}]interface{}", func() {
				expectedDirectorProperties := map[string]interface{}{
					"director_configuration": map[string]interface{}{
						"filtered_key": "filtered_key",
						"max_threads":  5,
						"encryption": map[interface{}]interface{}{
							"providers": map[interface{}]interface{}{
								"partition_password": "some_password",
								"client_certificate": "user_provided_cert",
								"client_key":         "user_provided_key",
								"client_user":        "user",
							},
						},
					},
				}
				fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  director_configuration:
    filtered_key: ((properties-configuration_director_configuration_filtered_key))
    encryption:
      providers:
        client_certificate: user_provided_cert
        client_key: ((properties-configuration_director_configuration_encryption_providers_client_key))
        client_user: ((properties-configuration_director_configuration_encryption_providers_client_user))
        partition_password: ((properties-configuration_director_configuration_encryption_providers_partition_password))
    max_threads: 5
vmtypes-configuration: {}
`)))
			})

			It("is able to handle []interface{}", func() {
				expectedDirectorProperties := map[string]interface{}{
					"iaas_configuration": map[string]interface{}{
						"filtered_key": "filtered_key",
						"encryption": map[string]interface{}{
							"providers": []interface{}{"some_key"},
						},
					},
				}
				fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties, nil)

				command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).ToNot(HaveOccurred())

				output := stdout.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_name: some-iaas
- name: some-other-az
  iaas_configuration_name: some-other-iaas
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: some-az
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: network-1
resource-configuration:
  some-job:
    additional_vm_extensions: ["some-vm-extension"]
    instances: 1
    instance_type:
      id: automatic
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
properties-configuration:
  iaas_configuration:
    filtered_key: ((properties-configuration_iaas_configuration_filtered_key))
    encryption:
      providers:
      - ((properties-configuration_iaas_configuration_encryption_providers_0))
vmtypes-configuration: {}
`)))
			})
		})

		Describe("failure cases", func() {
			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
				})
			})

			When("looking up the director GUID fails", func() {
				BeforeEach(func() {
					fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director properties fails", func() {
				BeforeEach(func() {
					fakeService.GetStagedDirectorPropertiesReturns(nil, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director azs fails", func() {
				BeforeEach(func() {
					fakeService.GetStagedDirectorAvailabilityZonesReturns(api.AvailabilityZonesOutput{}, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director networks fails", func() {
				BeforeEach(func() {
					fakeService.GetStagedDirectorNetworksReturns(api.NetworksConfigurationOutput{}, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director network assignment fails", func() {
				BeforeEach(func() {
					fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director jobs fails", func() {
				BeforeEach(func() {
					fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})

			When("looking up the director job resource config fails", func() {
				BeforeEach(func() {
					fakeService.ListStagedProductJobsReturns(map[string]string{
						"some-job": "some-job-guid",
					}, nil)
					fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
				})

				It("returns an error", func() {
					command := commands.NewStagedDirectorConfig(fakeService, stdout, stderr)
					err := command.Execute([]string{})
					Expect(err).To(MatchError("some-error"))
				})
			})
		})
	})
})
