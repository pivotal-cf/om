package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobType", func() {
	Context("HasPersistentDisk", func() {
		It("Should have persistent disk", func() {
			metadata := getMetadata("fixtures/metadata/pas.yml")
			job, err := metadata.GetJob("mysql")
			Expect(err).ToNot(HaveOccurred())
			Expect(job.HasPersistentDisk()).Should(BeTrue())
		})
		It("Should not have persistent disk", func() {
			metadata := getMetadata("fixtures/metadata/pas.yml")
			job, err := metadata.GetJob("router")
			Expect(err).ToNot(HaveOccurred())
			Expect(job.HasPersistentDisk()).Should(BeFalse())
		})
	})

	Context("GetPropertyBlueprint", func() {
		It("returns a configurable property", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			property, err := metadata.GetPropertyBlueprint(".healthwatch-forwarder.foundation_name")
			Expect(err).ToNot(HaveOccurred())
			Expect(property.Name).Should(Equal("foundation_name"))
		})
	})
})
