package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	var template proofing.Template

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		template = productTemplate.JobTypes[0].Templates[0]
	})

	It("parses their structure", func() {
		Expect(template.Consumes).To(Equal("some-consumes"))
		Expect(template.Manifest).To(Equal("some-manifest"))
		Expect(template.Name).To(Equal("some-name"))
		Expect(template.Provides).To(Equal("some-provides"))
		Expect(template.Release).To(Equal("some-release"))
	})
})
