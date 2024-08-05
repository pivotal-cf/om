package integration_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/configtemplate/generator"
)

func getMetadata(filename string) *generator.Metadata {
	fileData, err := os.ReadFile(filename)
	Expect(err).ToNot(HaveOccurred())
	metadata, err := generator.NewMetadata(fileData)
	Expect(err).ToNot(HaveOccurred())
	return metadata
}

var _ = Describe("Product PropertyInputs", func() {
	Context("GetAllProductProperties", func() {
		It("Should return new required product properties for healthwatch", func() {
			metadata := getMetadata("../fixtures/metadata/p_healthwatch.yml")
			expected, err := os.ReadFile("../fixtures/vars/healthwatch-product.yml")
			Expect(err).ToNot(HaveOccurred())
			productProperties, err := generator.GetAllProductProperties(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(productProperties).ToNot(BeNil())
			yml, err := yaml.Marshal(productProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})

		It("Should return new required product properties for pas", func() {
			metadata := getMetadata("../fixtures/metadata/pas.yml")
			expected, err := os.ReadFile("../fixtures/vars/pas-required.yml")
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
			metadata := getMetadata("../fixtures/metadata/p_healthwatch.yml")
			expected, err := os.ReadFile("../fixtures/vars/healthwatch-default-vars.yml")
			Expect(err).ToNot(HaveOccurred())
			requiredVars, err := generator.GetDefaultPropertyVars(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(requiredVars).ToNot(BeNil())
			yml, err := yaml.Marshal(requiredVars)
			Expect(err).ToNot(HaveOccurred())
			Expect(yml).Should(MatchYAML(string(expected)))
		})
	})

	Context("GetRequiredVars", func() {
		It("Should return required variables for properties that do not have defaults set", func() {
			expected := map[string]interface{}{
				"healthwatch-forwarder_health_check_az": "",
				"opsman_enable_url":                     "",
				"network_name":                          "",
				"singleton_availability_zone":           "",
				"service_network_name":                  "",
			}

			metadata := getMetadata("../fixtures/metadata/p_healthwatch.yml")
			requiredVars, err := generator.GetRequiredPropertyVars(metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(requiredVars).ToNot(BeNil())

			Expect(requiredVars).To(Equal(expected))
		})
	})
})
