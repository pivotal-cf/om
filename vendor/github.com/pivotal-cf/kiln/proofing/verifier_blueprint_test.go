package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VerifierBlueprint", func() {
	var verifierBlueprint proofing.VerifierBlueprint

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		verifierBlueprint = productTemplate.FormTypes[0].Verifiers[0]
	})

	It("parses their structure", func() {
		Expect(verifierBlueprint.Name).To(Equal("some-name"))
		Expect(verifierBlueprint.Properties).To(Equal("some-properties"))
	})
})
