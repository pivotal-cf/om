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

	runCommand := func() []interface{} {
		err := command.Execute([]string{
			"--product", "/path/to/a/product.pivotal",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(metadataExtractor.ExtractMetadataCallCount()).To(Equal(1))
		Expect(metadataExtractor.ExtractMetadataArgsForCall(0)).To(Equal("/path/to/a/product.pivotal"))

		Expect(logger.PrintlnCallCount()).To(Equal(1))
		output := logger.PrintlnArgsForCall(0)

		return output
	}

	BeforeEach(func() {
		logger = &fakes.Logger{}
		metadataExtractor = &fakes.MetadataExtractor{}
		command = commands.NewConfigTemplate(metadataExtractor, logger)
	})

	Describe("Execute", func() {
		It("writes a config file to output", func() {
			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Raw: []byte(`---
property_blueprints:
- name: some-string-property
  type: string
  optional: false
  configurable: true
- name: some-name
  type: boolean
  default: true
  optional: true
  configurable: true
`),
			}, nil)

			output := runCommand()
			Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-string-property:
    value: # required
  .properties.some-name:
    value: true`)))
		})
		Context("non-configurable property", func() {
			It("filters the property", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: boolean
  default: true
  configurable: false
- name: some-name1
  type: boolean
  default: true
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name1:
    value: true # required
`)))
			})
		})

		Context("optional property", func() {
			It("does not write '# required' next to the property", func() {
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

				output := runCommand()
				Expect(output).NotTo(ContainElement(ContainSubstring("# required")))
			})
		})

		Context("required property", func() {
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

				output := runCommand()
				lines := strings.Split(output[0].(string), "\n")
				valueOutput := strings.TrimSpace(lines[2])
				Expect(valueOutput).To(Equal("value: true # required"))
			})
		})

		Context("credential type", func() {
			It("write value string for secret type", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: secret
  optional: false
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: # required
      secret: ""
`)))
			})

			It("write value string for simple_credentials type", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: simple_credentials
  optional: false
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: # required
      identity: ""
      password: ""
`)))
			})

			It("write value string for rsa_cert_credentials type", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: rsa_cert_credentials
  optional: false
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: # required
      cert_pem: ""
      private_key_pem: ""
`)))
			})

			It("write value string for rsa_pkey_credentials type", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: rsa_pkey_credentials
  optional: false
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: # required
      public_key_pem: ""
      private_key_pem: ""
`)))
			})

			It("write value string for salted_credentials type", func() {
				metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
					Raw: []byte(`---
property_blueprints:
- name: some-name
  type: salted_credentials
  optional: false
  configurable: true
`),
				}, nil)

				output := runCommand()
				Expect(output).To(ContainElement(MatchYAML(`---
product-properties:
  .properties.some-name:
    value: # required
      identity: ""
      password: ""
      salt: ""
`)))
			})
		})

		Context("collection type", func() {
			Context("collection has default value", func() {
				It("prints default values as the inner value", func() {
					metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
						Raw: []byte(`---
property_blueprints:
- configurable: true
  name: some-name
  property_blueprints:
  - configurable: true
    default: false
    name: abc
    type: string
  - configurable: true
    name: def
    type: string
  type: collection
  default:
  - abc: a
    def: 1
  - abc: b
    def: 2
`),
					}, nil)

					output := runCommand()
					Expect(output).To(ContainElement(MatchYAML(`
product-properties:
  .properties.some-name:
    value: # required
    - abc: a
      def: 1
    - abc: b
      def: 2
`)))
				})
			})

			Context("inner key has default value", func() {
				It("prints default values as the inner value", func() {
					metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
						Raw: []byte(`---
property_blueprints:
- configurable: true
  name: some-name
  property_blueprints:
  - configurable: true
    default: false
    name: primary
    type: boolean
  type: collection
`),
					}, nil)

					output := runCommand()
					Expect(output).To(ContainElement(MatchYAML(`
product-properties:
  .properties.some-name:
    value: # required
    - primary: false
`)))
				})
			})

			Context("no default value", func() {
				It("prints null as the inner value", func() {
					metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
						Raw: []byte(`---
property_blueprints:
- configurable: true
  name: some-name
  property_blueprints:
  - configurable: true
    name: primary
    type: boolean
  type: collection
`),
					}, nil)

					output := runCommand()
					Expect(output).To(ContainElement(MatchYAML(`
product-properties:
  .properties.some-name:
    value: # required
    - primary: null
`)))
				})
			})
		})
	})

	Describe("with --include-placeholder flag", func() {
		It("replace credential types to placeholders", func() {
			metadataExtractor.ExtractMetadataReturns(extractor.Metadata{
				Raw: []byte(`---
property_blueprints:
- name: unrelated
  type: string
  default: some string
  optional: false
  configurable: true
- name: some-name
  type: simple_credentials
  optional: false
  configurable: true
- name: some-name1
  type: rsa_cert_credentials
  optional: false
  configurable: true
- name: some-name2
  type: rsa_pkey_credentials
  optional: false
  configurable: true
- name: some-name3
  type: salted_credentials
  optional: false
  configurable: true
- name: some-name4
  type: secret
  optional: false
  configurable: true
- name: some-name5
  type: collection
  optional: false
  property_blueprints:
  # credentials in collection
  - configurable: true
    name: key
    type: secret
  - configurable: true
    name: key1
    type: rsa_cert_credentials
  configurable: true
`),
			}, nil)

			output := runCommand()
			Expect(output).To(ContainElement(MatchYAML(`
product-properties:
  .properties.unrelated:
    value: some string
  .properties.some-name:
    value:
      identity: ""
      password: ""
  .properties.some-name1:
    value:
      cert_pem: ""
      private_key_pem: ""
  .properties.some-name2:
    value:
      private_key_pem: ""
      public_key_pem: ""
  .properties.some-name3:
    value:
      identity: ""
      password: ""
      salt: ""
  .properties.some-name4:
    value:
      secret: ""
  .properties.some-name5:
    value:
    - key:
        secret: ""
      key1:
        cert_pem: ""
        private_key_pem: ""
`)))
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "**EXPERIMENTAL** This command generates a configuration template that can be passed in to om configure-product",
				ShortDescription: "**EXPERIMENTAL** generates a config template for the product",
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
