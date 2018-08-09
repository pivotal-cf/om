package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CollectionPropertyInput", func() {
	var collectionPropertyInput proofing.CollectionPropertyInput

	BeforeEach(func() {
		f, err := os.Open("fixtures/form_types.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		var ok bool
		collectionPropertyInput, ok = productTemplate.FormTypes[0].PropertyInputs[1].(proofing.CollectionPropertyInput)
		Expect(ok).To(BeTrue())
	})

	It("parses their structure", func() {
		Expect(collectionPropertyInput.Description).To(Equal("some-description"))
		Expect(collectionPropertyInput.Label).To(Equal("some-label"))
		Expect(collectionPropertyInput.Placeholder).To(Equal("some-placeholder"))
		Expect(collectionPropertyInput.Reference).To(Equal("some-reference"))

		Expect(collectionPropertyInput.PropertyInputs).To(HaveLen(1))
	})

	Context("property_inputs", func() {
		It("parses their structure", func() {
			propertyInput := collectionPropertyInput.PropertyInputs[0]

			Expect(propertyInput.Description).To(Equal("some-description"))
			Expect(propertyInput.Label).To(Equal("some-label"))
			Expect(propertyInput.Placeholder).To(Equal("some-placeholder"))
			Expect(propertyInput.Reference).To(Equal("some-reference"))
			Expect(propertyInput.Slug).To(BeTrue())
		})
	})
})
