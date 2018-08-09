package proofing_test

import (
	"errors"
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PropertyBlueprints", func() {
	var productTemplate proofing.ProductTemplate

	BeforeEach(func() {
		f, err := os.Open("fixtures/property_blueprints.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err = proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())
	})

	It("parses the different types", func() {
		Expect(productTemplate.PropertyBlueprints[0]).To(BeAssignableToTypeOf(proofing.SimplePropertyBlueprint{}))
		Expect(productTemplate.PropertyBlueprints[1]).To(BeAssignableToTypeOf(proofing.SelectorPropertyBlueprint{}))
		Expect(productTemplate.PropertyBlueprints[2]).To(BeAssignableToTypeOf(proofing.CollectionPropertyBlueprint{}))
	})

	Context("failure cases", func() {
		Context("when the YAML cannot be unmarshalled", func() {
			It("returns an error", func() {
				propertyBlueprints := proofing.PropertyBlueprints([]proofing.PropertyBlueprint{})

				err := propertyBlueprints.UnmarshalYAML(func(v interface{}) error {
					return errors.New("unmarshal failed")
				})
				Expect(err).To(MatchError("unmarshal failed"))
			})
		})
	})
})
