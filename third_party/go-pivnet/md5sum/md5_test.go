package md5sum_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/v7/md5sum"
)

var _ = Describe("MD5", func() {
	Describe("FileSummer", func() {
		var (
			tempFilePath string
			tempDir      string
			fileContents []byte
			fileSummer *md5sum.FileSummer
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			fileContents = []byte("foobar contents")

			tempFilePath = filepath.Join(tempDir, "foobar")

			ioutil.WriteFile(tempFilePath, fileContents, os.ModePerm)

			fileSummer = md5sum.NewFileSummer()
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the MD5 of a file without error", func() {
			md5, err := fileSummer.SumFile(tempFilePath)
			Expect(err).NotTo(HaveOccurred())

			// Expected md5 of 'foobar contents'
			Expect(md5).To(Equal("fdd3d599138fd15d7673f3d3539531c1"))
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
