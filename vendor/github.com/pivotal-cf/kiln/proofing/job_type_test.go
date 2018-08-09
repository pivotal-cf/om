package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobType", func() {
	var jobType proofing.JobType

	BeforeEach(func() {
		f, err := os.Open("fixtures/job_types.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		jobType = productTemplate.JobTypes[0]
	})

	It("parses their structure", func() {
		Expect(jobType.Canaries).To(Equal(1))
		Expect(jobType.Description).To(Equal("some-description"))
		Expect(jobType.Errand).To(BeTrue())
		Expect(jobType.Manifest).To(Equal("some-manifest"))
		Expect(jobType.MaxInFlight).To(Equal("some-max-in-flight"))
		Expect(jobType.Name).To(Equal("some-name"))
		Expect(jobType.ResourceLabel).To(Equal("some-resource-label"))
		Expect(jobType.RunPreDeleteErrandDefault).To(BeTrue())
		Expect(jobType.RunPostDeployErrandDefault).To(BeTrue())
		Expect(jobType.Serial).To(BeTrue())
		Expect(jobType.SingleAZOnly).To(BeTrue())

		Expect(jobType.InstanceDefinition).To(BeAssignableToTypeOf(proofing.InstanceDefinition{}))
		Expect(jobType.PropertyBlueprints).To(HaveLen(1))
		Expect(jobType.ResourceDefinitions).To(HaveLen(1))
		Expect(jobType.Templates).To(HaveLen(1))
		Expect(jobType.RequiresProductVersions).To(HaveLen(1))
	})

	Context("property_blueprints", func() {
		It("parses their structure", func() {
			propertyBlueprint, ok := jobType.PropertyBlueprints[0].(proofing.SimplePropertyBlueprint)
			Expect(ok).To(BeTrue())

			Expect(propertyBlueprint.Configurable).To(BeTrue())
			Expect(propertyBlueprint.Constraints).To(Equal("some-constraints"))
			Expect(propertyBlueprint.Default).To(Equal("some-default"))
			Expect(propertyBlueprint.Name).To(Equal("some-name"))
			Expect(propertyBlueprint.Optional).To(BeTrue())
			Expect(propertyBlueprint.Type).To(Equal("some-type"))
		})
	})
})
