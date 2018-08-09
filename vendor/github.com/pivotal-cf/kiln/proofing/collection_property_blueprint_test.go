package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CollectionPropertyBlueprint", func() {
	var collectionPropertyBlueprint proofing.CollectionPropertyBlueprint

	BeforeEach(func() {
		f, err := os.Open("fixtures/property_blueprints.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		var ok bool
		collectionPropertyBlueprint, ok = productTemplate.PropertyBlueprints[2].(proofing.CollectionPropertyBlueprint)
		Expect(ok).To(BeTrue())
	})

	It("parses their structure", func() {
		Expect(collectionPropertyBlueprint.Name).To(Equal("some-collection-name"))
		Expect(collectionPropertyBlueprint.Type).To(Equal("collection"))
		Expect(collectionPropertyBlueprint.Default).To(Equal("some-default"))
		Expect(collectionPropertyBlueprint.Constraints).To(Equal("some-constraints"))
		Expect(collectionPropertyBlueprint.Configurable).To(BeTrue())
		Expect(collectionPropertyBlueprint.Optional).To(BeTrue())
		Expect(collectionPropertyBlueprint.FreezeOnDeploy).To(BeFalse())
		Expect(collectionPropertyBlueprint.Unique).To(BeFalse())
		Expect(collectionPropertyBlueprint.ResourceDefinitions).To(HaveLen(1))
	})

	Describe("Normalize", func() {
		It("returns a list of normalized property blueprints", func() {
			normalized := collectionPropertyBlueprint.Normalize("some-prefix")

			Expect(normalized).To(ConsistOf([]proofing.NormalizedPropertyBlueprint{
				{
					Property:     "some-prefix.some-collection-name",
					Configurable: true,
					Default:      "some-default",
					Required:     false,
					Type:         "collection",
				},
			}))
		})

		Context("when the property blueprint is not optional", func() {
			It("marks the property blueprint as required", func() {
				collectionPropertyBlueprint.Optional = false

				normalized := collectionPropertyBlueprint.Normalize("some-prefix")
				Expect(normalized[0].Required).To(BeTrue())
			})
		})
	})

})
