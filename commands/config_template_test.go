package commands_test

import (
	"errors"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"github.com/pivotal-cf/om/extractor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigTemplate", func() {
	var (
		logger            *fakes.Logger
		metadataExtractor *fakes.MetadataExtractor
		command           commands.ConfigTemplate
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		metadataExtractor = &fakes.MetadataExtractor{}
		metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
			Raw: []byte(`---
property_blueprints:
- name: some-name
  type: boolean
  default: true
  optional: true
  configurable: true
`),
		}, nil)

		command = commands.NewConfigTemplate(logger, metadataExtractor)
	})

	Describe("Execute", func() {
		It("writes a config file to output", func() {
			err := command.Execute([]string{
				"--product", "/path/to/a/product.pivotal",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
			Expect(metadataExtractor.ExtractMetadataArgsForCall(0)).To(Equal("/path/to/a/product.pivotal"))

			Expect(logger.PrintlnCallCount()).To(Equal(1))
			output := logger.PrintlnArgsForCall(0)
			Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: true`)))
		})

		Context("when a property is optional", func() {
			It("does not write '# required' next to the property", func() {
				err := command.Execute([]string{
					"--product", "/path/to/a/product.pivotal",
				})
				Expect(err).NotTo(HaveOccurred())

				output := logger.PrintlnArgsForCall(0)

				lines := strings.Split(output[0].(string), "\n")
				valueOutput := strings.TrimSpace(lines[2])
				Expect(valueOutput).To(Equal("value: true"))
			})
		})

		Context("when a property is not optional", func() {
			It("writes '# required' next to the property", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: boolean
  default: true
  optional: false
  configurable: true
`),
				}, nil)

				command = commands.NewConfigTemplate(logger, metadataExtractor)

				err := command.Execute([]string{
					"--product", "/path/to/a/product.pivotal",
				})
				Expect(err).NotTo(HaveOccurred())

				output := logger.PrintlnArgsForCall(0)

				lines := strings.Split(output[0].(string), "\n")
				valueOutput := strings.TrimSpace(lines[2])
				Expect(valueOutput).To(Equal("value: true # required"))
			})
		})

		Context("when the property is a simple credential", func() {
			FIt("prints out an identity and password field", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: simple_credentials
  optional: true
  configurable: true
`),
				}, nil)

				command = commands.NewConfigTemplate(logger, metadataExtractor)

				err := command.Execute([]string{
					"--product", "/path/to/a/product.pivotal",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
				Expect(metadataExtractor.ExtractMetadataArgsForCall(0)).To(Equal("/path/to/a/product.pivotal"))

				Expect(logger.PrintlnCallCount()).To(Equal(1))
				output := logger.PrintlnArgsForCall(0)
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value:
      identity:
      password:
`)))

			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a configuration template that can be passed in to om configure-product",
				ShortDescription: "generates a config template for the product",
				Flags:            command.Options,
			}))
		})
	})

	Describe("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when the product flag is not provided", func() {
			It("returns an error", func() {
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse config-template flags: missing required flag \"--product\""))
			})
		})

		Context("when the metadata cannot be extracted", func() {
			It("returns an error", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{}, errors.New("failed to extract"))

				err := command.Execute([]string{
					"--product", "/path/to/a/product.pivotal",
				})
				Expect(err).To(MatchError("could not extract metadata: failed to extract"))
			})
		})

		Context("when the metadata cannot be parsed", func() {
			It("returns an error", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte("%%%"),
				}, nil)

				err := command.Execute([]string{
					"--product", "/path/to/a/product.pivotal",
				})
				Expect(err).To(MatchError("could not parse metadata: yaml: could not find expected directive name"))
			})
		})
	})
})
