package proofing_test

import (
	"os"

	"github.com/pivotal-cf/kiln/proofing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Release", func() {
	var release proofing.Release

	BeforeEach(func() {
		f, err := os.Open("fixtures/metadata.yml")
		defer f.Close()
		Expect(err).NotTo(HaveOccurred())

		productTemplate, err := proofing.Parse(f)
		Expect(err).NotTo(HaveOccurred())

		release = productTemplate.Releases[0]
	})

	It("parses their structure", func() {
		Expect(release.File).To(Equal("some-file"))
		Expect(release.Name).To(Equal("some-name"))
		Expect(release.SHA1).To(Equal("some-sha1"))
		Expect(release.Version).To(Equal("some-version"))
	})

	It("is valid", func() {
		Expect(release.Validate()).To(Succeed())
	})

	Context("validations", func() {
		BeforeEach(func() {
			release = proofing.Release{
				Name:    "some-name",
				Version: "some-version",
				File:    "some-file",
				SHA1:    "some-sha1",
			}
		})

		It("validates the presence of the Name field", func() {
			release.Name = ""
			Expect(release.Validate()).To(MatchError("release name must be present"))
		})

		It("validates the presence of the File field", func() {
			release.File = ""
			Expect(release.Validate()).To(MatchError("release file must be present"))
		})

		It("validates the presence of the Name field", func() {
			release.Version = ""
			Expect(release.Validate()).To(MatchError("release version must be present"))
		})

		It("combines validations", func() {
			release = proofing.Release{}
			Expect(release.Validate()).To(MatchError(`- release name must be present
- release file must be present
- release version must be present`))
		})
	})
})
