package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("InstallationLog", func() {
	var (
		command     *commands.InstallationLog
		fakeService *fakes.InstallationLogService
		logger      *fakes.Logger
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		fakeService = &fakes.InstallationLogService{}
		command = commands.NewInstallationLog(fakeService, logger)
	})

	Describe("Execute", func() {
		It("displays the logs for the specified installation", func() {
			fakeService.GetInstallationLogsReturns(api.InstallationsServiceOutput{Logs: "some log output"}, nil)
			err := executeCommand(command, []string{
				"--id", "999",
			})

			Expect(err).ToNot(HaveOccurred())

			Expect(fakeService.GetInstallationLogsCallCount()).To(Equal(1))
			requestedInstallationId := fakeService.GetInstallationLogsArgsForCall(0)
			Expect(requestedInstallationId).To(Equal(999))

			Expect(logger.PrintCallCount()).To(Equal(1))
			outputLogs := logger.PrintArgsForCall(0)[0]
			Expect(outputLogs).To(Equal("some log output"))
		})
		When("the api fails to retrieve the installation log", func() {
			It("returns an error", func() {
				fakeService.GetInstallationLogsReturns(
					api.InstallationsServiceOutput{},
					errors.New("failed to retrieve installation log"),
				)
				err := executeCommand(command, []string{"--id", "999"})
				Expect(err).To(MatchError("failed to retrieve installation log"))
			})
		})
	})
})
