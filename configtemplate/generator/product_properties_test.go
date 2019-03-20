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
			Expect(err).ShouldNot(HaveOccurred())
			expected, err := ioutil.ReadFile("fixtures/healthwatch-required.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			productProperties, err := generator.CreateProductProperties(metadata)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(productProperties).ShouldNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})

		It("Should return new required product properties", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ShouldNot(HaveOccurred())
			expected, err := ioutil.ReadFile("fixtures/pas-required.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			productProperties, err := generator.CreateProductProperties(metadata)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(productProperties).ShouldNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})
	})

	Context("CreateProductPropertiesFeaturesOpsFiles", func() {
		It("Should return ops files map", func() {
			fileData, err := ioutil.ReadFile("fixtures/cloudcache.yml")
			Expect(err).ShouldNot(HaveOccurred())

			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			opsfiles, err := generator.CreateProductPropertiesFeaturesOpsFiles(metadata)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(opsfiles).ShouldNot(BeNil())
			// yml, err := yaml.Marshal(productProperties)
			// Expect(err).ShouldNot(HaveOccurred())
			// Expect(yml).Should(MatchYAML(string(expected)))
		})

	})
})
