package commands_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedConfig", func() {
	var (
		logger              *fakes.Logger
		stagedConfigService *fakes.StagedConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}

		stagedConfigService = &fakes.StagedConfigService{}
		stagedConfigService.PropertiesReturns(
			map[string]api.ResponseProperty{
				".properties.some-string-property": api.ResponseProperty{
					Value:        "some-value",
					Configurable: true,
				},
				".properties.some-non-configurable-property": api.ResponseProperty{
					Value:        "some-value",
					Configurable: false,
				},
				".properties.some-secret-property": api.ResponseProperty{
					Value: map[string]interface{}{
						"some-secret-type": "***",
					},
					IsCredential: true,
					Configurable: true,
				},
				".properties.some-null-property": api.ResponseProperty{
					Value:        nil,
					Configurable: true,
				},
			}, nil)
		stagedConfigService.NetworksAndAZsReturns(
			map[string]interface{}{
				"singleton_availability_zone": map[string]string{
					"name": "az-one",
				},
			}, nil)

		stagedConfigService.FindReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{
				GUID: "some-product-guid",
			},
		}, nil)

		stagedConfigService.JobsReturns(map[string]string{
			"some-job": "some-job-guid",
		}, nil)
		stagedConfigService.GetExistingJobConfigReturns(api.JobProperties{
			InstanceType: api.InstanceType{
				ID: "automatic",
			},
			Instances: 1,
		}, nil)
	})

	Describe("Execute", func() {
		It("writes a config file to output", func() {
			command := commands.NewStagedConfig(stagedConfigService, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(stagedConfigService.FindCallCount()).To(Equal(1))
			Expect(stagedConfigService.FindArgsForCall(0)).To(Equal("some-product"))

			Expect(stagedConfigService.PropertiesCallCount()).To(Equal(1))
			Expect(stagedConfigService.PropertiesArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(stagedConfigService.NetworksAndAZsCallCount()).To(Equal(1))
			Expect(stagedConfigService.NetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(stagedConfigService.JobsCallCount()).To(Equal(1))
			Expect(stagedConfigService.JobsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(stagedConfigService.GetExistingJobConfigCallCount()).To(Equal(1))
			productGuid, jobsGuid := stagedConfigService.GetExistingJobConfigArgsForCall(0)
			Expect(productGuid).To(Equal("some-product-guid"))
			Expect(jobsGuid).To(Equal("some-job-guid"))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-string-property:
    value: some-value
  .properties.some-secret-property:
    value:
      some-secret-type: "***"
network-properties:
  singleton_availability_zone:
    name: az-one
resource-config:
  some-job:
    instances: 1
    instance_type:
      id: automatic
`)))
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when product name is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse staged-config flags: missing required flag \"--product-name\""))
			})
		})

		Context("when looking up the product GUID fails", func() {
			BeforeEach(func() {
				stagedConfigService.FindReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the product properties fails", func() {
			BeforeEach(func() {
				stagedConfigService.PropertiesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the network fails", func() {
			BeforeEach(func() {
				stagedConfigService.NetworksAndAZsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing jobs fails", func() {
			BeforeEach(func() {
				stagedConfigService.JobsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the job fails", func() {
			BeforeEach(func() {
				stagedConfigService.GetExistingJobConfigReturns(api.JobProperties{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(stagedConfigService, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewStagedConfig(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
				ShortDescription: "generates a config from a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
