package generator_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/configtemplate/generator"
)

var _ = Describe("JobType", func() {
	Context("HasPersistentDisk", func() {
		It("Should have persistent disk", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			job, err := metadata.GetJob("mysql")
			Expect(err).ToNot(HaveOccurred())
			Expect(job.HasPersistentDisk()).Should(BeTrue())
		})
		It("Should not have persistent disk", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			job, err := metadata.GetJob("router")
			Expect(err).ToNot(HaveOccurred())
			Expect(job.HasPersistentDisk()).Should(BeFalse())
		})
	})

	Context("GetPropertyMetadata", func() {
		It("returns a configurable property", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ToNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ToNot(HaveOccurred())
			property, err := metadata.GetPropertyMetadata(".healthwatch-forwarder.foundation_name")
			Expect(err).ToNot(HaveOccurred())
			Expect(property.Name).Should(Equal("foundation_name"))
		})
	})
})
