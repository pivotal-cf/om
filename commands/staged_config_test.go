package commands_test

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/configparser"
	"reflect"
)

var _ = Describe("StagedConfig", func() {
	var (
		logger         *fakes.Logger
		fakeConfParser *fakes.ConfigParser
		fakeService    *fakes.StagedConfigService
	)

	valuePtr := func(t interface{}) interface{} {
		return reflect.ValueOf(t).Pointer()
	}

	BeforeEach(func() {
		logger = &fakes.Logger{}

		fakeConfParser = &fakes.ConfigParser{}

		fakeService = &fakes.StagedConfigService{}

		fakeService.GetStagedProductPropertiesReturns(
			map[string]api.ResponseProperty{
				".properties.some-string-property": {
					Value:        "some-value",
					Configurable: true,
				},
			}, nil,
		)

		fakeService.GetStagedProductNetworksAndAZsReturns(
			map[string]interface{}{
				"singleton_availability_zone": map[string]string{
					"name": "az-one",
				},
			}, nil)

		fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{
			Product: api.StagedProduct{
				GUID: "some-product-guid",
			},
		}, nil)

		fakeService.ListStagedProductJobsReturns(map[string]string{
			"some-job": "some-job-guid",
		}, nil)

		fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{
			InstanceType: api.InstanceType{
				ID: "automatic",
			},
			Instances: 1,
		}, nil)
	})

	Describe("Execute", func() {
		It("uses the default config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(valuePtr(handler)).To(Equal(valuePtr(configparser.NilHandler())))
		})

		It("writes a config file to stdout", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.GetStagedProductByNameCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductByNameArgsForCall(0)).To(Equal("some-product"))

			Expect(fakeService.GetStagedProductPropertiesCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductPropertiesArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.GetStagedProductNetworksAndAZsCallCount()).To(Equal(1))
			Expect(fakeService.GetStagedProductNetworksAndAZsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.ListStagedProductJobsCallCount()).To(Equal(1))
			Expect(fakeService.ListStagedProductJobsArgsForCall(0)).To(Equal("some-product-guid"))

			Expect(fakeService.GetStagedProductJobResourceConfigCallCount()).To(Equal(1))
			productGuid, jobsGuid := fakeService.GetStagedProductJobResourceConfigArgsForCall(0)
			Expect(productGuid).To(Equal("some-product-guid"))
			Expect(jobsGuid).To(Equal("some-job-guid"))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(ContainSubstring("product-properties:")))
			Expect(output).To(ContainElement(ContainSubstring("{}")))
			Expect(output).To(ContainElement(ContainSubstring("network-properties:")))
			Expect(output).To(ContainElement(ContainSubstring("az-one")))
			Expect(output).To(ContainElement(ContainSubstring("resource-config:")))
			Expect(output).To(ContainElement(ContainSubstring("some-job")))
		})
	})

	Context("when --include-placeholder is used", func() {
		It("uses the with placeholder config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-placeholder",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(valuePtr(handler)).To(Equal(valuePtr(configparser.PlaceholderHandler())))
		})

	})

	Context("when --include-credentials is used", func() {
		BeforeEach(func() {
			fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{
				{
					Type: "some-product",
					GUID: "some-product-guid",
				},
			}, nil)

			fakeService.GetDeployedProductCredentialReturns(api.GetDeployedProductCredentialOutput{
				Credential: api.Credential{
					Type: "some-secret-type",
					Value: map[string]string{
						"some-secret-key": "some-secret-value",
					},
				},
			}, nil)
		})

		It("uses the with credential config parser creds handler", func() {
			command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
			err := command.Execute([]string{
				"--product-name", "some-product",
				"--include-credentials",
			})
			Expect(err).NotTo(HaveOccurred())

			_, _, _, handler := fakeConfParser.ParsePropertiesArgsForCall(0)
			Expect(valuePtr(handler)).To(Equal(valuePtr(configparser.GetCredentialHandler(nil))))
		})

		Context("and the product has not yet been deployed", func() {
			BeforeEach(func() {
				fakeService.ListDeployedProductsReturns([]api.DeployedProductOutput{}, nil)
			})

			It("errors with a helpful message to the operator", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).To(MatchError("cannot retrieve credentials for product 'some-product': deploy the product and retry"))
			})
		})

		Context("and listing deployed products fails", func() {
			BeforeEach(func() {
				fakeService.ListDeployedProductsReturns(
					[]api.DeployedProductOutput{},
					errors.New("some-error"),
				)
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
					"--include-credentials",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse staged-config flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when product name is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse staged-config flags: missing required flag \"--product-name\""))
			})
		})

		Context("when looking up the product GUID fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductByNameReturns(api.StagedProductsFindOutput{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the product properties fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductPropertiesReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the network fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductNetworksAndAZsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when listing jobs fails", func() {
			BeforeEach(func() {
				fakeService.ListStagedProductJobsReturns(nil, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when looking up the job fails", func() {
			BeforeEach(func() {
				fakeService.GetStagedProductJobResourceConfigReturns(api.JobProperties{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				command := commands.NewStagedConfig(fakeService, fakeConfParser, logger)
				err := command.Execute([]string{
					"--product-name", "some-product",
				})
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {

			command := commands.NewStagedConfig(nil, nil, nil)

			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged product that can be passed in to om configure-product (Note: credentials are not available and will appear as '***')",
				ShortDescription: "**EXPERIMENTAL** generates a config from a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
