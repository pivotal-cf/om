package commands_test

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"io/ioutil"
)

var _ bool = Describe("StagedDirectorConfig", func() {
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
			expectedDirectorAZs := map[string][]map[string]interface{}{
				"availability_zones": {
					{
						"name": "some-az",
						"guid": "some-az-guid",
					},
				},
			}
			fakeService.GetStagedDirectorAvailabilityZonesReturns(expectedDirectorAZs, nil)

			expectedDirectorProperties := map[string]map[string]interface{}{
				"director_configuration": {
					"max_threads": 5,
				},
				"iaas_configuration": {
					"iaas_specific_key": "some-value",
				},
				"syslog_configuration": {
					"syslogconfig": "awesome",
				},
				"security_configuration": {
					"trusted_certificates": "some-certificate",
				},
			}
			fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties, nil)

			expectedNetworks := map[string]interface{}{
				"networks": []interface{}{
					map[string]string{
						"name": "network-1",
						"guid": "network-1-guid",
					},
				},
			}
			fakeService.GetStagedDirectorNetworksReturns(expectedNetworks, nil)

			fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
				api.StagedProduct{
					"p-bosh-guid",
					"director",
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
		})

		It("Writes a complete config file to output", func() {
			command := commands.NewStagedDirectorConfig(fakeService, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
  guid: some-az-guid
director-configuration:
  max_threads: 5
iaas-configuration:
  iaas_specific_key: some-value
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: "some-az"
networks-configuration:
  networks:
  - name: network-1
    guid: network-1-guid
resource-configuration:
  some-job:
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
`)))
		})

		It("writes the config to a file", func() {
			outFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			command := commands.NewStagedDirectorConfig(fakeService, logger)
			err = command.Execute([]string{
				"-o", outFile.Name(),
			})
			Expect(err).NotTo(HaveOccurred())

			output, err := ioutil.ReadFile(outFile.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(string(output)).To(MatchYAML(`
az-configuration:
- name: some-az
  guid: some-az-guid
director-configuration:
  max_threads: 5
iaas-configuration:
  iaas_specific_key: some-value
network-assignment:
  network:
    name: network-1
  singleton_availability_zone:
    name: "some-az"
networks-configuration:
  networks:
  - name: network-1
    guid: network-1-guid
resource-configuration:
  some-job:
    instances: 1
    instance_type:
      id: automatic
security-configuration:
  trusted_certificates: some-certificate
syslog-configuration:
  syslogconfig: awesome
`))
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
				fakeService.GetStagedDirectorAvailabilityZonesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedDirectorConfig(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the director networks fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedDirectorNetworksReturns(nil, errors.New("some-error"))
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
