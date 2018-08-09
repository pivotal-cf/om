package proofing_test

import (
	"errors"
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PropertyInputs", func() {
	var formType proofing.FormType

	BeforeEach(func() {
		f, err := os.Open("fixtures/form_types.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		formType = productTemplate.FormTypes[0]
	})

	It("parses the different types", func() {
		Expect(formType.PropertyInputs[0]).To(BeAssignableToTypeOf(proofing.SimplePropertyInput{}))
		Expect(formType.PropertyInputs[1]).To(BeAssignableToTypeOf(proofing.CollectionPropertyInput{}))
		Expect(formType.PropertyInputs[2]).To(BeAssignableToTypeOf(proofing.SelectorPropertyInput{}))
	})

	Context("failure cases", func() {
		Context("when the YAML cannot be unmarshalled", func() {
			It("returns an error", func() {
				propertyInputs := proofing.PropertyInputs([]proofing.PropertyInput{})

				err := propertyInputs.UnmarshalYAML(func(v interface{}) error {
					return errors.New("unmarshal failed")
				})

				Expect(err).To(MatchError("unmarshal failed"))
			})
		})
	})
})
