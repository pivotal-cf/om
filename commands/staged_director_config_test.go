package commands_test

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("StagedDirectorConfig", func() {
	var (
		logger      *fakes.Logger
		fakeService *fakes.StagedDirectorConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		fakeService = &fakes.StagedDirectorConfigService{}
	})

	Describe("Execute", func() {

		BeforeEach(func() {
			expectedDirectorAZs := api.AvailabilityZonesOutput{
				AvailabilityZones: []api.AvailabilityZoneOutput{
					{
						Name:                  "some-az",
						IAASConfigurationGUID: "some-iaas-guid",
					},
					{
						Name: "some-other-az",
					},
				},
			}
			fakeService.GetStagedDirectorAvailabilityZonesReturns(expectedDirectorAZs, nil)

			expectedDirectorProperties := map[string]map[string]interface{}{
				"director_configuration": {
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
				"iaas_configuration": {
					"project": "project-id",
					"key":     "some-key",
				},
				"syslog_configuration": {
					"syslogconfig": "awesome",
				},
				"security_configuration": {
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
				Instances: 1,
				InstanceType: api.InstanceType{
					ID: "automatic",
				},
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
		})

		It("Writes a complete config file with filtered sensitive fields to stdout", func() {
			command := commands.NewStagedDirectorConfig(fakeService, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_guid: some-iaas-guid
- name: some-other-az
director-configuration:
  max_threads: 5
  encryption:
    providers:
      client_certificate: user_provided_cert
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
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
`)))
		})

		Describe("when getting availability_zones returns an empty array", func() {
			It("doesn't return the az in the config", func() {
				fakeService.GetStagedDirectorAvailabilityZonesReturns(api.AvailabilityZonesOutput{}, nil)
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				output := logger.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
director-configuration:
  max_threads: 5
  encryption:
    providers:
      client_certificate: user_provided_cert
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
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
`)))
			})
		})

		Describe("with --include-credentials", func() {
			It("Includes the filtered fields when printing to stdout", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{
					"--include-credentials",
				})
				Expect(err).NotTo(HaveOccurred())

				output := logger.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_guid: some-iaas-guid
- name: some-other-az
director-configuration:
  filtered_key: filtered_key
  max_threads: 5
  encryption:
    providers:
      client_certificate: user_provided_cert
      partition_password: some_password
      client_key: user_provided_key
      client_user: user
iaas-configuration:
  key: some-key
  project: project-id
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
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
`)))
			})
		})

		Describe("with --include-placeholders", func() {
			It("Includes the placeholder fields when printing to stdout", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{
					"--include-placeholders",
				})
				Expect(err).NotTo(HaveOccurred())

				output := logger.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  iaas_configuration_guid: some-iaas-guid
- name: some-other-az
director-configuration:
  filtered_key: ((director-configuration_filtered_key))
  encryption:
    providers:
      client_certificate: user_provided_cert
      client_key: ((director-configuration_encryption_providers_client_key))
      client_user: ((director-configuration_encryption_providers_client_user))
      partition_password: ((director-configuration_encryption_providers_partition_password))
  max_threads: 5
iaas-configuration:
  project: ((iaas-configuration_project))
  key: ((iaas-configuration_key))
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
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
vmextensions-configuration:
  - name: vm_ext1
    cloud_properties: 
      source_dest_check: false
  - name: vm_ext2
    cloud_properties:
      key_name: operations_keypair
`)))
			})
		})

	})

	Describe("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when looking up the director GUID fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director properties fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedDirectorPropertiesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director azs fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedDirectorAvailabilityZonesReturns(api.AvailabilityZonesOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director networks fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedDirectorNetworksReturns(api.NetworksConfigurationOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director network assignment fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director jobs fails", func() {
			BeforeEach(func() {
				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director job resource config fails", func() {
			BeforeEach(func() {
				fakeService.ListStagedProductJobsReturns(map[string]string{
					"some-job": "some-job-guid",
				}, nil)
				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStagedDirectorConfig(nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged director that can be passed in to om configure-director",
				ShortDescription: "**EXPERIMENTAL** generates a config from a staged director",
				Flags:            command.Options,
			}))
		})
	})
})
