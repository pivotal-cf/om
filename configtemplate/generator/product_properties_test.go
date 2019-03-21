package generator_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Product Properties", func() {
	Context("CreateProductProperties", func() {
		It("Should return new required product properties", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile("fixtures/healthwatch-required.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			productProperties, err := generator.CreateProductProperties(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(productProperties).ToNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})

		It("Should return new required product properties", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile("fixtures/pas-required.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			productProperties, err := generator.CreateProductProperties(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(productProperties).ToNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})
	})

	Context("CreateProductPropertiesFeaturesOpsFiles", func() {
		When("there is a property that is a selector", func() {
			It("returns the value and selected value", func() {
				metadata := &generator.Metadata{
					PropertyMetadata: []generator.PropertyMetadata{
						{
							Type: "selector",
							Name: "some_selector",
							OptionTemplates: []generator.OptionTemplate{
								{
									Name:        "gcp",
									SelectValue: "GCP",
								},
							},
						},
					},
					FormTypes: []generator.FormType{
						{
							Properties: []generator.Property{
								{
									Reference: ".properties.some_selector",
									Selectors: []generator.SelectorProperty{
										{
											Reference: ".properties.some_selector.gcp",
										},
									},
								},
							},
						},
					},
				}
				opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(opsfiles["some_selector-gcp"]).To(ContainElement(generator.Ops{
					Type: "replace",
					Path: "/product-properties/.properties.some_selector?",
					Value: &generator.OpsValue{
						Value:          "GCP",
						SelectedOption: "gcp",
					},
				}))
			})
		})
	})
})
