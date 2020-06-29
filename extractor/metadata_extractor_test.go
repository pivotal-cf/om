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
name: some-product
stemcell_criteria:
  os: ubuntu-trusty
  enable_patch_security_updates: true
  version: "3586"
`
)

func createProductFile(metadataFilePath, contents string) *os.File {
	var err error
	productFile, err := ioutil.TempFile("", "")
	Expect(err).ToNot(HaveOccurred())

	stat, err := productFile.Stat()
	Expect(err).ToNot(HaveOccurred())

	zipper := zip.NewWriter(productFile)

	productWriter, err := zipper.CreateHeader(&zip.FileHeader{
		Name:               metadataFilePath,
		UncompressedSize64: uint64(stat.Size()),
		Modified:           stat.ModTime(),
	})
	Expect(err).ToNot(HaveOccurred())

	_, err = io.WriteString(productWriter, contents)
	Expect(err).ToNot(HaveOccurred())

	err = zipper.Close()
	Expect(err).ToNot(HaveOccurred())

	return productFile
}

var _ = Describe("MetadataExtractor", func() {
	var (
		metadataExtractor extractor.MetadataExtractor
		productFile       *os.File
	)

	BeforeEach(func() {
		productFile = createProductFile("metadata/some-product.yml", validYAML)
		metadataExtractor = extractor.MetadataExtractor{}
	})

	AfterEach(func() {
		os.Remove(productFile.Name())
	})

	Describe("ExtractFromFile", func() {
		It("Extracts the product name and version from the given pivotal file", func() {
			metadata, err := metadataExtractor.ExtractFromFile(productFile.Name())
			Expect(err).ToNot(HaveOccurred())

			Expect(metadata.Name).To(Equal("some-product"))
			Expect(metadata.Version).To(Equal("1.8.14"))
			Expect(metadata.StemcellCriteria.OS).To(Equal("ubuntu-trusty"))
			Expect(metadata.StemcellCriteria.Version).To(Equal("3586"))
			Expect(metadata.StemcellCriteria.PatchSecurityUpdates).To(BeTrue())
			Expect(metadata.Raw).To(MatchYAML(validYAML))
		})

		When("an error occurs", func() {
			When("the product tarball does not exist", func() {
				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile("fake-file")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			When("no metadata file is found", func() {
				var badProductFile *os.File
				BeforeEach(func() {
					badProductFile = createProductFile("", validYAML)
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile(badProductFile.Name())
					Expect(err).To(MatchError("no metadata file was found in provided .pivotal"))
				})
			})

			When("the metadata file contains bad YAML", func() {
				var badProductFile *os.File

				BeforeEach(func() {
					badProductFile = createProductFile("./metadata/some-product.yml", `%%%`)
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile(badProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("could not extract product metadata: yaml: could not find expected directive name")))
				})
			})

			When("the metadata file does not contain product name or version", func() {
				var badProductFile *os.File

				BeforeEach(func() {
					badProductFile = createProductFile("./metadata/some-product.yml", `foo: bar`)
				})

				AfterEach(func() {
					os.Remove(badProductFile.Name())
				})

				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile(badProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("could not extract product metadata: could not find product details in metadata file")))
				})
			})

			When("the metadata file is in the wrong place", func() {
				var wrongProductFile *os.File

				BeforeEach(func() {
					wrongProductFile = createProductFile("some-product.yml", validYAML)
				})

				AfterEach(func() {
					os.Remove(wrongProductFile.Name())
				})

				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile(wrongProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("no metadata file was found in provided .pivotal")))
				})
			})

			When("the metadata file is in a subdirectory", func() {
				var nestedProductFile *os.File
				BeforeEach(func() {
					nestedProductFile = createProductFile("__MACOSX/metadata/._metadata.yml", validYAML)
				})

				AfterEach(func() {
					os.Remove(nestedProductFile.Name())
				})

				It("returns an error", func() {
					_, err := metadataExtractor.ExtractFromFile(nestedProductFile.Name())
					Expect(err).To(MatchError(ContainSubstring("no metadata file was found in provided .pivotal")))
				})
			})
		})
	})
})
