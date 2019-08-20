package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExportInstallation", func() {
	var (
		fakeService *fakes.ExportInstallationService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.ExportInstallationService{}
		logger = &fakes.Logger{}
	})

	It("exports the installation", func() {
		command := commands.NewExportInstallation(fakeService, logger)

		err := command.Execute([]string{
			"--output-file", "/path/to/output.zip",
		})
		Expect(err).NotTo(HaveOccurred())

		By("calling export on the installation service")
		Expect(fakeService.DownloadInstallationAssetCollectionCallCount()).To(Equal(1))
		outputFile := fakeService.DownloadInstallationAssetCollectionArgsForCall(0)
		Expect(outputFile).To(Equal("/path/to/output.zip"))

		By("printing correct log messages")
		Expect(logger.PrintfCallCount()).To(Equal(2))
		format, v := logger.PrintfArgsForCall(0)
		Expect(fmt.Sprintf(format, v...)).To(Equal("exporting installation"))

		format, v = logger.PrintfArgsForCall(1)
		Expect(fmt.Sprintf(format, v...)).To(Equal("finished exporting installation"))
	})

	Context("failure cases", func() {
		When("an unknown flag is provided", func() {
			It("returns an error", func() {
				command := commands.NewExportInstallation(fakeService, logger)
				err := command.Execute([]string{"--badflag"})
				Expect(err).To(MatchError("could not parse export-installation flags: flag provided but not defined: -badflag"))
			})
		})

		When("output file is not provided", func() {
			It("returns an error and prints out usage", func() {
				command := commands.NewExportInstallation(fakeService, logger)
				err := command.Execute([]string{})
				Expect(err).To(MatchError("could not parse export-installation flags: missing required flag \"--output-file\""))
			})
		})

		When("the installation cannot be exported", func() {
			It("returns an error", func() {
				command := commands.NewExportInstallation(fakeService, logger)
				fakeService.DownloadInstallationAssetCollectionReturns(errors.New("some error"))

				err := command.Execute([]string{"--output-file", "/some/path"})
				Expect(err).To(MatchError("failed to export installation: some error"))
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewExportInstallation(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command will export the current installation of the target Ops Manager.",
				ShortDescription: "exports the installation of the target Ops Manager",
				Flags:            command.Options,
			}))
		})
	})
})
