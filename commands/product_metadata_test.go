package commands_test

import (
	"archive/zip"

	"os"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ProductMetadata", func() {
	Describe("Execute", func() {
		var (
			command commands.ProductMetadata
			stdout  *fakes.Logger

			productFile *os.File
			err         error
		)

		BeforeEach(func() {
			stdout = &fakes.Logger{}

			command = commands.NewProductMetadata(stdout)

			// write fake file
			productFile, err = ioutil.TempFile("", "fake-tile")
			z := zip.NewWriter(productFile)

			// https://github.com/pivotal-cf/om/issues/239
			// writing a "directory" as well, because some tiles seem to
			// have this as a separate file in the zip, which influences the regexp
			// needed to capture the metadata file
			_, err := z.Create("metadata/")
			Expect(err).ToNot(HaveOccurred())

			f, err := z.Create("metadata/fake-tile.yml")
			Expect(err).ToNot(HaveOccurred())

			_, err = f.Write([]byte(`
name: fake-tile
product_version: 1.2.3
`))
			Expect(err).ToNot(HaveOccurred())

			Expect(z.Close()).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(productFile.Name())).To(Succeed())
		})

		It("shows product name from tile metadata file", func() {
			err = command.Execute([]string{
				"-p",
				productFile.Name(),
				"--product-name",
			})
			Expect(err).ToNot(HaveOccurred())

			content := stdout.PrintlnArgsForCall(0)
			Expect(content).To(ContainElement("fake-tile"))
		})

		It("shows product version from tile metadata file", func() {
			err = command.Execute([]string{
				"-p",
				productFile.Name(),
				"--product-version",
			})
			Expect(err).ToNot(HaveOccurred())

			content := stdout.PrintlnArgsForCall(0)
			Expect(content).To(ContainElement("1.2.3"))
		})

		Context("failure cases", func() {
			When("the flags cannot be parsed", func() {
				It("returns an error", func() {
					err = command.Execute([]string{"--bad-flag", "some-value"})
					Expect(err).To(MatchError(MatchRegexp("could not parse product-metadata flags")))
				})
			})

			When("the flags are not specified", func() {
				It("returns an error", func() {
					err = command.Execute([]string{"-p", productFile.Name()})
					Expect(err).To(MatchError(MatchRegexp("you must specify product-name and/or product-version")))
				})
			})

			When("the specified product file is not found", func() {
				It("returns an error", func() {
					err = command.Execute([]string{"-p", "non-existent-file", "--product-name"})
					Expect(err).To(MatchError(MatchRegexp("failed to open product file")))
				})
			})

			When("the file does not have metadata", func() {
				var (
					badTile *os.File
				)

				BeforeEach(func() {
					badTile, err = ioutil.TempFile("", "bad-tile")
					Expect(err).ToNot(HaveOccurred())
					z := zip.NewWriter(badTile)
					Expect(z.Close()).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(badTile.Name())).To(Succeed())
				})

				It("returns an error", func() {
					err = command.Execute([]string{"-p", badTile.Name(), "--product-name"})
					Expect(err).To(MatchError(MatchRegexp("failed to find metadata file")))
				})
			})
		})
	})

	Describe("Usage", func() {
		var (
			command commands.ProductMetadata
			stdout  *fakes.Logger
		)

		BeforeEach(func() {
			stdout = &fakes.Logger{}
		})

		It("returns the usage information for the product-metadata command", func() {
			command = commands.NewProductMetadata(stdout)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command prints metadata about the given product",
				ShortDescription: "prints product metadata",
				Flags:            command.Options,
			}))
		})

		It("returns the usage information for the tile-metadata command", func() {
			command = commands.NewDeprecatedProductMetadata(stdout)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "***DEPRECATED*** use 'product-metadata' instead\nThis command prints metadata about the given product",
				ShortDescription: "**DEPRECATED** prints product metadata. Use product-metadata instead",
				Flags:            command.Options,
			}))
		})
	})
})
