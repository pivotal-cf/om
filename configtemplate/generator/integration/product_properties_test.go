package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var _ = Describe("Product PropertyInputs", func() {
	Context("GetAllProductProperties", func() {
		It("Should return new required product properties", func() {
			fileData, err := ioutil.ReadFile("../fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile("../fixtures/healthwatch-product.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			productProperties, err := generator.GetAllProductProperties(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(productProperties).ToNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})

		It("Should return new required product properties", func() {
			fileData, err := ioutil.ReadFile("../fixtures/pas.yml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile("../fixtures/pas-required.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			productProperties, err := generator.GetAllProductProperties(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(productProperties).ToNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})
	})

	Context("GetDefaultVars", func() {
		It("Should return default variables for properties", func() {
			fileData, err := ioutil.ReadFile("../fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile("../fixtures/healthwatch-default-vars.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			requiredVars, err := generator.GetDefaultPropertyVars(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(requiredVars).ToNot(BeNil())
			yml, err := yaml.Marshal(requiredVars)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})
	})
})
