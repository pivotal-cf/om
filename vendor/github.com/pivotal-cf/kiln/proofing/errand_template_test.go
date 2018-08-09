package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ErrandTemplate", func() {
	var errandTemplate proofing.ErrandTemplate

	BeforeEach(func() {
		f, err := os.Open("fixtures/errands.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		errandTemplate = productTemplate.PostDeployErrands[0]
	})

	It("parses their structure", func() {
		Expect(errandTemplate.Colocated).To(BeTrue())
		Expect(errandTemplate.Description).To(Equal("some-description"))
		Expect(errandTemplate.Instances).To(HaveLen(1))
		Expect(errandTemplate.Label).To(Equal("some-label"))
		Expect(errandTemplate.Name).To(Equal("some-name"))
		Expect(errandTemplate.RunDefault).To(BeTrue())
	})
})
