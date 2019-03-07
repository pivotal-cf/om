package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
)

var _ = Describe("ConfigTemplate", func() {
	var (
		//logger  *fakes.Logger
		command *commands.ConfigTemplate
	)

	BeforeEach(func() {
		command = commands.NewConfigTemplate()
	})

	Describe("Execute", func() {
		PDescribe("upserting an entry in the output directory with template files", func() {
			When("the output directory does not exist", func() {
				It("returns an error indicating the path does not exist", func() {
				})

			})
			When("the output directory exists with the product's directory", func() {
				It("updates the files in the product's sub-directory", func() {
				})
			})
			When("the output directory already exists without the product's directory", func() {
				It("creates a new subdirectory named the by the product version", func() {
				})
				It("creates files in the new subdirectory", func() {

				})
			})
		})
	})

	PWhen("there are non-configurable properties", func() {
		It("does not include them in product.yml", func() {
		})

		When("there are optional properties", func() {
			It("does not include them in product.yml", func() {
			})
			It("creates an ops-file adding that property in the options directory", func() {
			})
		})

		When("there are required properties", func() {
			It("includes the property in the product yaml config", func() {
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command generates a product configuration template from a .pivotal file on Pivnet",
				ShortDescription: "generates a config template from a Pivnet product",
				Flags:            command.Options,
			}))
		})
	})

	Describe("flag handling", func() {
		When("an unknown flag is provided", func() {
			It("returns an error", func() {
				err := command.Execute([]string{"--invalid"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -invalid"))
				err = command.Execute([]string{"--unreal"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -unreal"))
			})
		})

		When("the output-directory flag is not provided", func() {
			var args []string
			BeforeEach(func() {
				args = []string{
					"",
				}
			})
			It("returns an error", func() {
				err := command.Execute(args)
				Expect(err).To(MatchError("could not parse config-template flags: missing required flag \"--output-directory\""))
			})
		})

		PWhen("the base-directory flag is not provided", func() {
			It("returns an error", func() {

			})
		})

		PDescribe("metadata extraction and parsing failures", func() {
			PWhen("the metadata cannot be extracted", func() {
				It("returns an error", func() {
				})
			})

			PWhen("the metadata cannot be parsed", func() {
				It("returns an error", func() {
				})
			})
		})
	})
})
