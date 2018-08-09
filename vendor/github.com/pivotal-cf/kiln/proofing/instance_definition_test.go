package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceDefinition", func() {
	var instanceDefinition proofing.InstanceDefinition

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		instanceDefinition = productTemplate.JobTypes[0].InstanceDefinition
	})

	It("parses their structure", func() {
		Expect(instanceDefinition.Configurable).To(BeTrue())
		Expect(instanceDefinition.Default).To(Equal(2))
		Expect(instanceDefinition.ZeroIf.PropertyReference).To(Equal("some-property-reference"))
		Expect(instanceDefinition.Constraints).To(Equal("some-constraints"))
	})
})
