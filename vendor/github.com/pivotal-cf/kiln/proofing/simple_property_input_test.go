package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SimplePropertyInput", func() {
	var simplePropertyInput proofing.SimplePropertyInput

	BeforeEach(func() {
		f, err := os.Open("fixtures/form_types.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		var ok bool
		simplePropertyInput, ok = productTemplate.FormTypes[0].PropertyInputs[0].(proofing.SimplePropertyInput)
		Expect(ok).To(BeTrue())
	})

	It("parses their structure", func() {
		Expect(simplePropertyInput.Description).To(Equal("some-description"))
		Expect(simplePropertyInput.Label).To(Equal("some-label"))
		Expect(simplePropertyInput.Placeholder).To(Equal("some-placeholder"))
		Expect(simplePropertyInput.Reference).To(Equal("some-reference"))
	})
})
