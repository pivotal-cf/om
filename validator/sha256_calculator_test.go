package validator_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pivotal-cf/om/validator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileSHA256HashCalculator", func() {
	var (
		fileSHA256HashCalculator validator.FileSHA256HashCalculator
		fileToSHA256             string
	)

	BeforeEach(func() {
		fileSHA256HashCalculator = validator.NewSHA256Calculator()

		tempDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		fileToSHA256 = filepath.Join(tempDir, "file-to-sum")
		err = ioutil.WriteFile(fileToSHA256, []byte("file contents"), 0644)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Checksum", func() {
		It("Calculates the checksum", func() {
			md5, err := fileSHA256HashCalculator.Checksum(fileToSHA256)
			Expect(err).ToNot(HaveOccurred())

			Expect(md5).To(Equal("7bb6f9f7a47a63e684925af3608c059edcc371eb81188c48c9714896fb1091fd"))
		})

		When("the file cannot be read", func() {
			It("returns an error", func() {
				_, err := fileSHA256HashCalculator.Checksum("non-existent-file")
				Expect(err).To(MatchError("open non-existent-file: no such file or directory"))
			})
		})
	})
})
