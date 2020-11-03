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
		fakeService *fakes.ExportInstallationService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		fakeService = &fakes.ExportInstallationService{}
		logger = &fakes.Logger{}
	})

	It("exports the installation", func() {
		command := commands.NewExportInstallation(fakeService, logger)

		err := executeCommand(command, []string{
			"--output-file", "/path/to/output.zip",
		})
		Expect(err).ToNot(HaveOccurred())

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

	When("the installation cannot be exported", func() {
		It("returns an error", func() {
			command := commands.NewExportInstallation(fakeService, logger)
			fakeService.DownloadInstallationAssetCollectionReturns(errors.New("some error"))

			err := executeCommand(command, []string{"--output-file", "/some/path"})
			Expect(err).To(MatchError("failed to export installation: some error"))
		})
	})
})
