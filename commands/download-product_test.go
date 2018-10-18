package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = FDescribe("DownloadProduct", func() {
	var (
		logger *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
	})

	PIt("downloads a product from Pivotal Network", func() {
		//command := commands.NewDownloadProduct(logger)

	})

	Context("failure cases", func() {
		Context("when an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewDownloadProduct(logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse download-product flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when a required flag is not provided", func() {
			It("returns an error", func() {
				command := commands.NewDownloadProduct(logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse download-product flags: missing required flag \"--pivnet-api-token\""))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDownloadProduct(nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
				ShortDescription: "downloads a specified product file from Pivotal Network",
				Flags:            command.Options,
			}))
		})
	})
})
