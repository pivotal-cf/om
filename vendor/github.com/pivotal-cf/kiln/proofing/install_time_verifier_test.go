package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstallTimeVerifier", func() {
	var installTimeVerifier proofing.InstallTimeVerifier

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		installTimeVerifier = productTemplate.InstallTimeVerifiers[0]
	})

	It("parses their structure", func() {
		Expect(installTimeVerifier.Ignorable).To(BeTrue())
		Expect(installTimeVerifier.Name).To(Equal("some-name"))

		Expect(installTimeVerifier.Properties).To(Equal("some-properties"))
	})
})
