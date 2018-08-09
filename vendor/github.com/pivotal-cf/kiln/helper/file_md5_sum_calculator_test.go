package helper_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pivotal-cf/kiln/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileMD5SumCalculator", func() {
	var (
		fileMD5SumCalculator helper.FileMD5SumCalculator
		fileToMD5            string
	)

	BeforeEach(func() {
		fileMD5SumCalculator = helper.NewFileMD5SumCalculator()

		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		fileToMD5 = filepath.Join(tempDir, "file-to-sum")
		err = ioutil.WriteFile(fileToMD5, []byte("file contents"), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Checksum", func() {
		It("Calculates the checksum", func() {
			md5, err := fileMD5SumCalculator.Checksum(fileToMD5)
			Expect(err).NotTo(HaveOccurred())

			Expect(md5).To(Equal("4a8ec4fa5f01b4ab1a0ab8cbccb709f0"))
		})

		Context("when the file cannot be read", func() {
			It("returns an error", func() {
				_, err := fileMD5SumCalculator.Checksum("non-existent-file")
				Expect(err).To(MatchError("open non-existent-file: no such file or directory"))
			})
		})
	})
})
