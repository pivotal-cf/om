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
		It("Should return ops files map", func() {
			fileData, err := ioutil.ReadFile("fixtures/cloudcache.yml")
			Expect(err).ToNot(HaveOccurred())

			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(opsfiles).ToNot(BeNil())
			// yml, err := yaml.Marshal(productProperties)
			// Expect(err).ToNot(HaveOccurred())
			// Expect(yml).Should(MatchYAML(string(expected)))
		})

	})
})
