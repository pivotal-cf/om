package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourceDefinitions", func() {
	var resourceDefinition proofing.ResourceDefinition

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		resourceDefinition = productTemplate.JobTypes[0].ResourceDefinitions[0]
	})

	It("parses their structure", func() {
		Expect(resourceDefinition.Configurable).To(BeTrue())
		Expect(resourceDefinition.Constraints).To(Equal("some-constraints"))
		Expect(resourceDefinition.Default).To(Equal(1))
		Expect(resourceDefinition.Name).To(Equal("some-name"))
	})
})
