package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

var _ bool = FDescribe("StagedDirectorConfig", func() {
	var (
		logger      *fakes.Logger
		fakeService *fakes.StagedDirectorConfigService
		expectedDirectorAZ []interface{}


	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		fakeService = &fakes.StagedDirectorConfigService{}



	})

	Describe("Execute", func() {

		It("Writes a complete config file to output", func() {
			expectedDirectorAZ = []interface{}{map[string]interface{}{"name": "some-az"}}
			fakeService.GetStagedDirectorAZReturns(expectedDirectorAZ, nil)

			expectedDirectorProperties := map[string]interface{}{"max_threads": 5}
			fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties,nil)

			command := commands.NewStagedDirectorConfig(fakeService, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`
az-configuration:
- name: some-az
director-configuration:
 max_threads: 5
iaas-configuration:
 iaas_specific_key: some-value
network-assignment:
 network:
   name: some-network
networks-configuration:
 networks:
 - network: network-1
resource-configuration:
 compilation:
   instance_type:
     id: m4.xlarge
security-configuration:
 trusted_certificates: some-certificate
syslog-configuration:
 syslogconfig: awesome
`)))

		})
		It("writes the AZ configured on Ops Man in the config file", func() {
			fakeService = &fakes.StagedDirectorConfigService{}
			expectedDirectorAZ = []interface{}{map[string]interface{}{"name": "exp-az"}}
			fakeService.GetStagedDirectorAZReturns(expectedDirectorAZ,nil)
			command := commands.NewStagedDirectorConfig(fakeService, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := map[string]interface{}{}
			err = yaml.Unmarshal([]byte(strings.Trim(fmt.Sprint(logger.PrintlnArgsForCall(0)), "[]")), &output)
			Expect(err).NotTo(HaveOccurred())
			returned, err :=  yaml.Marshal(output["az-configuration"])
			Expect(err).NotTo(HaveOccurred())

			expected, err := yaml.Marshal(expectedDirectorAZ)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(returned)).To(Equal(string(expected)))
		})
		It("writes the director-configuration configured on Ops Man in the config file", func() {
			fakeService = &fakes.StagedDirectorConfigService{}
			expectedDirectorProperties := map[string]interface{}{"max_threads": 10}
			fakeService.GetStagedDirectorPropertiesReturns(expectedDirectorProperties,nil)
			command := commands.NewStagedDirectorConfig(fakeService, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := map[string]interface{}{}
			err = yaml.Unmarshal([]byte(strings.Trim(fmt.Sprint(logger.PrintlnArgsForCall(0)), "[]")), &output)
			Expect(err).NotTo(HaveOccurred())
			returned, err :=  yaml.Marshal(output["director-configuration"])
			Expect(err).NotTo(HaveOccurred())

			expected, err := yaml.Marshal(expectedDirectorProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(returned)).To(Equal(string(expected)))
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

		Context("when looking up the product properties fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
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
