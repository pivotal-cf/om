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

var _ = Describe("ExportTemplate", func() {
	var (
		logger              *fakes.Logger
		exportConfigService *fakes.ExportConfigService
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		exportConfigService = &fakes.ExportConfigService{}
	})

	Describe("Execute", func() {
		It("writes a config file to output", func() {
			command := commands.NewExportConfig(exportConfigService, logger)
			exportConfigService.ExportConfigReturns(api.ExportConfigOutput{
				Properties: map[string]api.OutputProperty{
					".properties.some-string-property": api.OutputProperty{
						Value: "some-value",
					},
				},
				NetworkProperties: map[string]interface{}{
					"singleton_availability_zone": map[string]string{
						"name": "az-one",
					},
				},
			}, nil)

			err := command.Execute([]string{
				"--product-name", "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(exportConfigService.ExportConfigCallCount()).To(Equal(1))
			Expect(exportConfigService.ExportConfigArgsForCall(0)).To(Equal("some-product"))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-string-property:
    value: some-value
network-properties:
  singleton_availability_zone:
    name: az-one
`)))
		})

	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewExportConfig(exportConfigService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse export-config flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when product name is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewExportConfig(exportConfigService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse export-config flags: missing required flag \"--product-name\""))
			})
		})

		Context("when the config cannot be exported", func() {
			It("returns an error", func() {
				command := commands.NewExportConfig(exportConfigService, logger)
				exportConfigService.ExportConfigReturns(api.ExportConfigOutput{}, errors.New("some error"))

				err := command.Execute([]string{"--product-name", "some-product"})
				Expect(err).To(MatchError("failed to export config: some error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewExportConfig(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a config from a staged product that can be passed in to om configure-product",
				ShortDescription: "generates a config from a staged product",
				Flags:            command.Options,
			}))
		})
	})
})
