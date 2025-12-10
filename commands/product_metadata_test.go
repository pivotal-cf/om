package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("ProductMetadata", func() {
	Describe("Execute", func() {
		var (
			command *commands.ProductMetadata
			stdout  *fakes.Logger
			err     error
		)

		BeforeEach(func() {
			stdout = &fakes.Logger{}

			command = commands.NewProductMetadata(func(*commands.ProductMetadata) (commands.MetadataProvider, error) {
				f := &fakes.MetadataProvider{}
				f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
				return f, nil
			}, stdout)
		})

		It("shows product name from tile metadata file", func() {
			err = executeCommand(command, []string{
				"-p",
				"product-filename",
				"--product-name",
			})
			Expect(err).ToNot(HaveOccurred())

			content := stdout.PrintlnArgsForCall(0)
			Expect(content).To(ContainElement("example-product"))
		})

		It("shows product version from tile metadata file", func() {
			err = executeCommand(command, []string{
				"-p",
				"product-filename",
				"--product-version",
			})
			Expect(err).ToNot(HaveOccurred())

			content := stdout.PrintlnArgsForCall(0)
			Expect(content).To(ContainElement("1.1.1"))
		})

		Describe("flag handling", func() {
			When("the required flags are not specified", func() {
				It("returns an error", func() {
					err = executeCommand(command, []string{"-p", "product-filename"})
					Expect(err).To(MatchError(MatchRegexp("you must specify product-name and/or product-version")))
				})
			})

			When("pivnet and product path args are provided", func() {
				It("returns an error", func() {
					err := executeCommand(command, []string{
						"--pivnet-api-token", "b",
						"--product-path", "c",
						"--product-name",
					})
					Expect(err).To(MatchError(ContainSubstring("please provide either pivnet flags OR product-path")))
				})
			})
		})

		When("the specified product file is not found", func() {
			BeforeEach(func() {
				command = commands.NewProductMetadata(func(*commands.ProductMetadata) (commands.MetadataProvider, error) {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns(nil, errors.New("open non-existent-file"))
					return f, nil
				}, stdout)
			})

			It("returns an error", func() {
				err = executeCommand(command, []string{"-p", "non-existent-file", "--product-name"})
				Expect(err).To(MatchError(MatchRegexp("open non-existent-file")))
			})
		})
	})
})
