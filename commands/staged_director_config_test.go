package commands_test

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ bool = FDescribe("StagedDirectorConfig", func() {
	var (
		logger      *fakes.Logger
		fakeService *fakes.StagedConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
	})

	Describe("Execute", func() {
		It("Writes a config file to output", func() {
			command := commands.NewStagedDirectorConfig(fakeService, logger)
			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			output := logger.PrintlnArgsForCall(0)

			Expect(output).To(ContainElement(MatchYAML(`---
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
	})
	//
	//	BeforeEach(func() {
	//		logger = &fakes.Logger{}
	//
	//		fakeService = &fakes.StagedConfigService{}
	//		fakeService.GetStagedProductPropertiesReturns(
	//			map[string]api.ResponseProperty{
	//				".properties.some-string-property": api.ResponseProperty{
	//					Value:        "some-value",
	//					Configurable: true,
	//				},
	//				".properties.some-non-configurable-property": api.ResponseProperty{
	//					Value:        "some-value",
	//					Configurable: false,
	//				},
	//				".properties.some-secret-property": api.ResponseProperty{
	//					Value: map[string]interface{}{
	//						"some-secret-type": "***",
	//					},
	//					IsCredential: true,
	//					Configurable: true,
	//				},
	//				".properties.some-non-configurable-secret-property": api.ResponseProperty{
	//					Value: map[string]interface{}{
	//						"some-secret-type": "***",
	//					},
	//					IsCredential: true,
	//					Configurable: false,
	//				},
	//				".properties.some-null-property": api.ResponseProperty{
	//					Value:        nil,
	//					Configurable: true,
	//				},
	//			}, nil)
	//		fakeService.GetStagedProductNetworksAndAZsReturns(
	//			map[string]interface{}{
	//				"singleton_availability_zone": map[string]string{
	//					"name": "az-one",
	//				},
	//			}, nil)
	//
	//		fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
	//			Product: api.StagedProduct{
	//				GUID: "some-product-guid",
	//			},
	//		}, nil)
	//
	//		fakeService.ListStagedProductJobsReturns(map[string]string{
	//			"some-job": "some-job-guid",
	//		}, nil)
	//		fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{
	//			InstanceType: api.InstanceType{
	//				ID: "automatic",
	//			},
	//			Instances: 1,
	//		}, nil)
	//	})
	//
	//	Describe("Execute", func() {
	//		It("writes a config file to output", func() {
	//			command := commands.NewStagedConfig(fakeService, logger)
	//			err := command.Execute([]string{
	//				"--product-name", "some-product",
	//			})
	//			Expect(err).NotTo(HaveOccurred())
	//
	//			Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
	//			Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))
	//
	//			Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
	//			Expect(fakeService.GetStagedProductPropertiesArgsForCall(0)).To(Equal("some-product-guid"))
	//
	//			Expect(fakeService.GetStagedProductNetworksAndAZsCallCount()).To(Equal(1))
	//			Expect(fakeService.GetStagedProductNetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))
	//
	//			Expect(fakeService.ListStagedProductJobsCallCount()).To(Equal(1))
	//			Expect(fakeService.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))
	//
	//			Expect(fakeService.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
	//			productGuid, jobsGuid := fakeService.GetStagedProductJobResourceConfigArgsForCall(0)
	//			Expect(productGuid).To(Equal("some-product-guid"))
	//			Expect(jobsGuid).To(Equal("some-job-guid"))
	//
	//			Expect(logger.PrintlnCallCount()).To(Equal(1))
	//			output := logger.PrintlnArgsForCall(0)
	//			Expect(output).To(ContainElement(MatchYAML(`---
	//product-properties:
	//  .properties.some-string-property:
	//    value: some-value
	//  .properties.some-secret-property:
	//    value:
	//      some-secret-type: "***"
	//network-properties:
	//  singleton_availability_zone:
	//    name: az-one
	//resource-config:
	//  some-job:
	//    instances: 1
	//    instance_type:
	//      id: automatic
	//`)))
	//		})
	//	})
	//
	//	Context("when --include-credentials is used", func() {
	//		BeforeEach(func() {
	//			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
	//				{
	//					Type: "some-product",
	//					GUID: "some-product-guid",
	//				},
	//			}, nil)
	//
	//			fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
	//				Credential: api.Credential{
	//					Type: "some-secret-type",
	//					Value: map[string]string{
	//						"some-secret-key": "some-secret-value",
	//					},
	//				},
	//			}, nil)
	//		})
	//
	//		It("includes secret values in the output", func() {
	//			command := commands.NewStagedConfig(fakeService, logger)
	//			err := command.Execute([]string{
	//				"--product-name", "some-product",
	//				"--include-credentials",
	//			})
	//			Expect(err).NotTo(HaveOccurred())
	//
	//			Expect(fakeService.GetDeployedProductCredentialCallCount()).To(Equal(1))
	//			Expect(fakeService.GetDeployedProductCredentialArgsForCall(0)).To(Equal(api.GetDeployedProductCredentialInput{
	//				DeployedGUID:        "some-product-guid",
	//				CredentialReference: ".properties.some-secret-property",
	//			}))
	//
	//			Expect(logger.PrintlnCallCount()).To(Equal(1))
	//			output := logger.PrintlnArgsForCall(0)
	//			Expect(output).To(ContainElement(MatchYAML(`---
	//product-properties:
	//  .properties.some-string-property:
	//    value: some-value
	//  .properties.some-secret-property:
	//    value:
	//      some-secret-key: some-secret-value
	//network-properties:
	//  singleton_availability_zone:
	//    name: az-one
	//resource-config:
	//  some-job:
	//    instances: 1
	//    instance_type:
	//      id: automatic
	//`)))
	//		})
	//
	//		Context("and the product has not yet been deployed", func() {
	//			BeforeEach(func() {
	//				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
	//			})
	//			It("errors with a helpful message to the operator", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//					"--include-credentials",
	//				})
	//				Expect(err).To(MatchError("cannot retrieve credentials for product 'some-product': deploy the product and retry"))
	//			})
	//		})
	//
	//		Context("and listing deployed products fails", func() {
	//			BeforeEach(func() {
	//				fakeService.ListDeployedProductsReturns(
	//					[]api.DeployedProductOutput{},
	//					errors.New("some-error"),
	//				)
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//					"--include-credentials",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//		Context("and looking up a credential fails", func() {
	//			BeforeEach(func() {
	//				fakeService.GetDeployedProductCredentialReturns(
	//					api.GetDeployedProductCredentialOutput{},
	//					errors.New("some-error"),
	//				)
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//					"--include-credentials",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//	})
	//
	//	Context("failure cases", func() {
	//		Context("when an unknown flag is provided", func() {
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{"--badflag"})
	//				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
	//			})
	//		})
	//
	//		Context("when product name is not provided", func() {
	//			It("returns an error and prints out usage", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{})
	//				Expect(err).To(MatchError("could not parse staged-config flags: missing required flag \"--product-name\""))
	//			})
	//		})
	//
	//		Context("when looking up the product GUID fails", func() {
	//			BeforeEach(func() {
	//				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//		Context("when looking up the product properties fails", func() {
	//			BeforeEach(func() {
	//				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//		Context("when looking up the network fails", func() {
	//			BeforeEach(func() {
	//				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//		Context("when listing jobs fails", func() {
	//			BeforeEach(func() {
	//				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//		Context("when looking up the job fails", func() {
	//			BeforeEach(func() {
	//				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
	//			})
	//
	//			It("returns an error", func() {
	//				command := commands.NewStagedConfig(fakeService, logger)
	//				err := command.Execute([]string{
	//					"--product-name", "some-product",
	//				})
	//				Expect(err).To(MatchError("some-error"))
	//			})
	//		})
	//
	//	})

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
