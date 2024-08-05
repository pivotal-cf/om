package generator_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/configtemplate/generator"
)

func getMetadata(filename string) *generator.Metadata {
	fileData, err := os.ReadFile(filename)
	Expect(err).ToNot(HaveOccurred())
	metadata, err := generator.NewMetadata(fileData)
	Expect(err).ToNot(HaveOccurred())
	return metadata
}

var _ = Describe("Metadata", func() {
	Context("UsesServiceNetwork", func() {
		It("Should use service network", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			Expect(metadata.UsesServiceNetwork()).Should(BeTrue())
		})

		It("Should not service network", func() {
			metadata := getMetadata("fixtures/metadata/pas.yml")
			Expect(metadata.UsesServiceNetwork()).Should(BeFalse())
		})
	})

	Context("GetPropertyBlueprint", func() {
		It("returns a non-job configurable property", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			property, err := metadata.GetPropertyBlueprint(".properties.opsman")
			Expect(err).ToNot(HaveOccurred())
			Expect(property.Name).Should(Equal("opsman"))
		})

		It("returns a job configurable property", func() {
			metadata := getMetadata("fixtures/metadata/p_healthwatch.yml")
			property, err := metadata.GetPropertyBlueprint(".healthwatch-forwarder.foundation_name")
			Expect(err).ToNot(HaveOccurred())
			Expect(property).ToNot(BeNil())
			Expect(property.Name).Should(Equal("foundation_name"))
		})
	})

	DescribeTable("ProductName tile metadata fixture tests", func(fixtureFilepath string, expectedName string) {
		metadata := getMetadata(fixtureFilepath)
		Expect(metadata.ProductName()).Should(BeEquivalentTo(expectedName))
	},
		Entry("PAS", "fixtures/metadata/pas.yml", "cf"),
		Entry("healthwatch", "fixtures/metadata/p_healthwatch.yml", "p-healthwatch"),
		Entry("iso-segment", "fixtures/metadata/iso-segment.yml", "p-isolation-segment"),
		Entry("replicated iso-segment", "fixtures/metadata/iso-segment-replicator.yml", "p-isolation-segment-new-seg"),
	)

	DescribeTable("ProductVersion tile metadata fixture tests", func(fixtureFilepath string, expectedVersion string) {
		metadata := getMetadata(fixtureFilepath)
		Expect(metadata.ProductVersion()).Should(BeEquivalentTo(expectedVersion))
	},
		Entry("PAS", "fixtures/metadata/pas.yml", "2.1.3"),
		Entry("healthwatch", "fixtures/metadata/p_healthwatch.yml", "1.2.1-build.1"),
		Entry("iso-segment", "fixtures/metadata/iso-segment.yml", "2.2.4"),
		Entry("replicated iso-segment", "fixtures/metadata/iso-segment-replicator.yml", "2.2.4"),
	)
})
