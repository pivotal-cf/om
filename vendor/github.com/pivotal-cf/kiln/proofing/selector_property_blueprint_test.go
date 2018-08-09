package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SelectorPropertyBlueprint", func() {
	var selectorPropertyBlueprint proofing.SelectorPropertyBlueprint

	BeforeEach(func() {
		f, err := os.Open("fixtures/property_blueprints.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		var ok bool
		selectorPropertyBlueprint, ok = productTemplate.PropertyBlueprints[1].(proofing.SelectorPropertyBlueprint)
		Expect(ok).To(BeTrue())
	})

	It("parses their structure", func() {
		Expect(selectorPropertyBlueprint.Configurable).To(BeTrue())
		Expect(selectorPropertyBlueprint.Constraints).To(Equal("some-constraints"))
		Expect(selectorPropertyBlueprint.Default).To(Equal("some-default"))
		Expect(selectorPropertyBlueprint.FreezeOnDeploy).To(BeTrue())
		Expect(selectorPropertyBlueprint.Name).To(Equal("some-selector-name"))
		Expect(selectorPropertyBlueprint.Optional).To(BeTrue())
		Expect(selectorPropertyBlueprint.Type).To(Equal("selector"))
		Expect(selectorPropertyBlueprint.Unique).To(BeTrue())
		Expect(selectorPropertyBlueprint.ResourceDefinitions).To(HaveLen(1))
		Expect(selectorPropertyBlueprint.OptionTemplates).To(HaveLen(1))
	})

	Describe("Normalize", func() {
		It("returns a list of normalized property blueprints", func() {
			normalized := selectorPropertyBlueprint.Normalize("some-prefix")

			Expect(normalized).To(ConsistOf([]proofing.NormalizedPropertyBlueprint{
				{
					Property:     "some-prefix.some-selector-name",
					Configurable: true,
					Default:      "some-default",
					Required:     false,
					Type:         "selector",
				},
				{
					Property:     "some-prefix.some-selector-name.some-option-template-name.some-nested-simple-name",
					Configurable: true,
					Default:      1,
					Required:     false,
					Type:         "some-type",
				},
			}))
		})

		Context("when the property blueprint is not optional", func() {
			It("marks the property blueprint as required", func() {
				selectorPropertyBlueprint.Optional = false

				normalized := selectorPropertyBlueprint.Normalize("some-prefix")
				Expect(normalized[0].Required).To(BeTrue())
			})
		})
	})
})
