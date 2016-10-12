package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExportInstallation", func() {
	var (
		installationService *fakes.InstallationService
		logger              *fakes.OtherLogger
	)

	BeforeEach(func() {
		installationService = &fakes.InstallationService{}
		logger = &fakes.OtherLogger{}
	})

	It("exports the installation", func() {
		command := commands.NewExportInstallation(installationService, logger)

		err := command.Execute([]string{
			"--output-file", "/path/to/output.zip",
		})
		Expect(err).NotTo(HaveOccurred())

		By("calling export on the installation service")
		Expect(installationService.ExportCallCount()).To(Equal(1))
		Expect(installationService.ExportArgsForCall(0)).To(Equal("/path/to/output.zip"))

		By("printing correct log messages")
		Expect(logger.PrintfCallCount()).To(Equal(2))
		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("exporting installation"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished exporting installation"))
	})

	Context("failure cases", func() {
		Context("when an unkwown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewExportInstallation(installationService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse export-installation flags: flag provided but not defined: -badflag"))
			})
		})

		Context("when output file is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewExportInstallation(installationService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("expected flag --output-file. Run 'om help export-installation' for more information."))
			})
		})

		Context("when the installation cannot be exported", func() {
			It("returns and error", func() {
				command := commands.NewExportInstallation(installationService, logger)
				installationService.ExportReturns(errors.New("some error"))

				err := command.Execute([]string{"--output-file", "/some/path"})
				Expect(err).To(MatchError("failed to export installation: some error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewExportInstallation(nil, nil)
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This command will export the current installation of the target Ops Manager.",
				ShortDescription: "exports the installation of the target ops manager",
				Flags:            command.Options,
			}))
		})
	})
})
