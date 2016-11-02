package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const productProperties = `{
  ".properties.something": {"value": "configure-me"},
  ".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} }
}`

const networkProperties = `{
  "singleton_availability_zone": {"name": "az-one"},
  "other_availability_zones": [{"name": "az-two" }, {"name": "az-three"}],
  "network": {"name": "network-one"}
}`

var _ = Describe("ConfigureProduct", func() {
	Describe("Execute", func() {
		var (
			service *fakes.ProductConfigurer
			logger  *fakes.Logger
		)

		BeforeEach(func() {
			service = &fakes.ProductConfigurer{}
			logger = &fakes.Logger{}
		})

		It("configures a product", func() {
			client := commands.NewConfigureProduct(service, logger)

			service.StagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "some-product-guid", Type: "cf"},
					{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
				},
			}, nil)

			err := client.Execute([]string{
				"--product-name", "cf",
				"--product-properties", productProperties,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.StagedProductsCallCount()).To(Equal(1))
			Expect(service.ConfigureArgsForCall(0)).To(Equal(api.ProductsConfigurationInput{
				GUID:          "some-product-guid",
				Configuration: productProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))
		})

		It("configures a product with product-network flag", func() {
			client := commands.NewConfigureProduct(service, logger)

			service.StagedProductsReturns(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{GUID: "some-product-guid", Type: "cf"},
					{GUID: "not-the-guid-you-are-looking-for", Type: "something-else"},
				},
			}, nil)

			err := client.Execute([]string{
				"--product-name", "cf",
				"--product-properties", productProperties,
				"--product-network", networkProperties,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.StagedProductsCallCount()).To(Equal(1))
			Expect(service.ConfigureArgsForCall(0)).To(Equal(api.ProductsConfigurationInput{
				GUID:          "some-product-guid",
				Configuration: productProperties,
				Network:       networkProperties,
			}))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("setting properties"))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("finished setting properties"))
		})

		Context("failure cases", func() {
			Context("when an unknown flag is provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(service, logger)
					err := command.Execute([]string{"--badflag"})
					Expect(err).To(MatchError("could not parse configure-product flags: flag provided but not defined: -badflag"))
				})
			})

			Context("when the product-name flag is not provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(service, logger)
					err := command.Execute([]string{"--product-properties", "{}"})
					Expect(err).To(MatchError("error: product-name is missing. Please see usage for more information."))
				})
			})

			Context("when the product-properties flag is not provided", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(service, logger)
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("error: product-properties or network-properties are required. Please see usage for more information."))
				})
			})

			Context("when the network-properties flag is empty", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(service, logger)
					err := command.Execute([]string{"--product-name", "some-product"})
					Expect(err).To(MatchError("error: product-properties or network-properties are required. Please see usage for more information."))
				})
			})

			Context("when the product cannot be configured", func() {
				It("returns an error", func() {
					command := commands.NewConfigureProduct(service, logger)
					service.ConfigureReturns(errors.New("some product error"))

					err := command.Execute([]string{"--product-name", "some-product", "--product-properties", "{}", "--product-network", "anything"})
					Expect(err).To(MatchError("failed to configure product: some product error"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewConfigureProduct(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command configures a staged product",
				ShortDescription: "configures a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
