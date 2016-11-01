package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	commonfakes "github.com/pivotal-cf/om/common/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const properties = `{
  ".properties.something": {"value": "configure-me"},
  ".a-job.job-property": {"value": {"identity": "username", "password": "example-new-password"} }
}`

var _ = Describe("ConfigureProduct", func() {
	Describe("Execute", func() {
		var (
			service *fakes.ProductConfigurer
			logger  *commonfakes.OtherLogger
		)

		BeforeEach(func() {
			service = &fakes.ProductConfigurer{}
			logger = &commonfakes.OtherLogger{}
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
				"--product-properties", properties,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.StagedProductsCallCount()).To(Equal(1))
			Expect(service.ConfigureArgsForCall(0)).To(Equal(api.ProductsConfigurationInput{
				GUID:          "some-product-guid",
				Configuration: properties,
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
					Expect(err).To(MatchError("error: product-properties is missing. Please see usage for more information."))
				})
			})

			Context("when the product cannot be configured", func() {
				It("returns and error", func() {
					command := commands.NewConfigureProduct(service, logger)
					service.ConfigureReturns(errors.New("some product error"))

					err := command.Execute([]string{"--product-name", "some-product", "--product-properties", "{}"})
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
