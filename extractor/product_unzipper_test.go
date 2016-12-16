package extractor_test

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/extractor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	validYAML = `
---
product_version: 1.8.14
name: some-product`
)

var _ = Describe("Product Unzipper", func() {
	var (
		unzipper    extractor.ProductUnzipper
		productFile *os.File
	)

	BeforeEach(func() {
		var err error
		productFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		stat, err := productFile.Stat()
		Expect(err).NotTo(HaveOccurred())

		zipper := zip.NewWriter(productFile)

		productWriter, err := zipper.CreateHeader(&zip.FileHeader{
			Name:               "./metadata/some-product.yml",
			UncompressedSize64: uint64(stat.Size()),
			ModifiedTime:       uint16(stat.ModTime().Unix()),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = io.WriteString(productWriter, validYAML)
		Expect(err).NotTo(HaveOccurred())

		err = zipper.Close()
		Expect(err).NotTo(HaveOccurred())

		unzipper = extractor.ProductUnzipper{}
	})

	AfterEach(func() {
		os.Remove(productFile.Name())
	})

	Describe("ExtractMetadata", func() {
		It("Extracts the product name and version from the given pivotal file", func() {
			name, version, err := unzipper.ExtractMetadata(productFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(name).To(Equal("some-product"))
			Expect(version).To(Equal("1.8.14"))
		})

		Context("when an error occurs", func() {
			Context("when the product tarball does not exist", func() {
				It("returns an error", func() {
					_, _, err := unzipper.ExtractMetadata("fake-file")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when no metadata file is found", func() {
				var badProductFile *os.File
				BeforeEach(func() {
					var err error
					badProductFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					zipper := zip.NewWriter(badProductFile)

					err = zipper.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, _, err := unzipper.ExtractMetadata(badProductFile.Name())
					Expect(err).To(MatchError("no metadata file was found in provided .pivotal"))
				})
			})

			Context("when the metadata file contains bad YAML", func() {
				var badProductFile *os.File

				BeforeEach(func() {
					var err error
					badProductFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())
					stat, err := badProductFile.Stat()
					Expect(err).NotTo(HaveOccurred())

					zipper := zip.NewWriter(badProductFile)
					productWriter, err := zipper.CreateHeader(&zip.FileHeader{
						Name:               "./metadata/some-product.yml",
						UncompressedSize64: uint64(stat.Size()),
						ModifiedTime:       uint16(stat.ModTime().Unix()),
					})
					Expect(err).NotTo(HaveOccurred())

					_, err = io.WriteString(productWriter, `%%%`)
					Expect(err).NotTo(HaveOccurred())

					err = zipper.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, _, err := unzipper.ExtractMetadata(badProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("could not extract product metadata: yaml: could not find expected directive name")))
				})
			})

			Context("when the metadata file does not contain product name or version", func() {
				var badProductFile *os.File

				BeforeEach(func() {
					var err error
					badProductFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					stat, err := badProductFile.Stat()
					Expect(err).NotTo(HaveOccurred())

					zipper := zip.NewWriter(badProductFile)
					productWriter, err := zipper.CreateHeader(&zip.FileHeader{
						Name:               "./metadata/some-product.yml",
						UncompressedSize64: uint64(stat.Size()),
						ModifiedTime:       uint16(stat.ModTime().Unix()),
					})
					Expect(err).NotTo(HaveOccurred())

					_, err = io.WriteString(productWriter, `foo: bar`)
					Expect(err).NotTo(HaveOccurred())

					err = zipper.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, _, err := unzipper.ExtractMetadata(badProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("could not extract product metadata: could not find product details in metadata file")))
				})
			})

			Context("when the metadata file is in the wrong place", func() {
				var wrongProductFile *os.File

				BeforeEach(func() {
					var err error
					wrongProductFile, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					stat, err := wrongProductFile.Stat()
					Expect(err).NotTo(HaveOccurred())

					zipper := zip.NewWriter(wrongProductFile)
					productWriter, err := zipper.CreateHeader(&zip.FileHeader{
						Name:               "some-product.yml",
						UncompressedSize64: uint64(stat.Size()),
						ModifiedTime:       uint16(stat.ModTime().Unix()),
					})
					Expect(err).NotTo(HaveOccurred())

					_, err = io.WriteString(productWriter, validYAML)
					Expect(err).NotTo(HaveOccurred())

					err = zipper.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.Remove(wrongProductFile.Name())
				})

				It("returns an error", func() {
					_, _, err := unzipper.ExtractMetadata(wrongProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("no metadata file was found in provided .pivotal")))
				})
			})
		})
	})
})
