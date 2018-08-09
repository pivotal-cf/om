package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Variable", func() {
	var variable proofing.Variable

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		variable = productTemplate.Variables[0]
	})

	It("parses their structure", func() {
		Expect(variable.Name).To(Equal("some-name"))
		Expect(variable.Options).To(Equal("some-options"))
		Expect(variable.Type).To(Equal("some-type"))
	})
})
