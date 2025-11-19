package sha256sum_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v7/sha256sum"
)

var _ = Describe("SHA256", func() {
	Describe("FileSummer", func() {
		var (
			tempFilePath string
			tempDir      string
			fileContents []byte

			fileSummer *sha256sum.FileSummer
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			fileContents = []byte("foobar contents")

			tempFilePath = filepath.Join(tempDir, "foobar")

			ioutil.WriteFile(tempFilePath, fileContents, os.ModePerm)

			fileSummer = sha256sum.NewFileSummer()
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the SHA256 of a file without error", func() {
			sha256, err := fileSummer.SumFile(tempFilePath)
			Expect(err).NotTo(HaveOccurred())

			// Expected sha256 of 'foobar contents'
			Expect(sha256).To(Equal("070a103eb906d53a5933d96f3301635d6c416491d6a0ebd0bf4d4e448af5762d"))
		})

		Context("when there is an error reading the file", func() {
			BeforeEach(func() {
				tempFilePath = "/not/a/valid/file"
			})

			It("returns the error", func() {
				_, err := fileSummer.SumFile(tempFilePath)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
