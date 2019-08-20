package commands_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
)

var _ = Describe("InstallationLog", func() {
	var (
		command     commands.InstallationLog
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
			err := command.Execute([]string{
				"--id", "999",
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeService.GetInstallationLogsCallCount()).To(Equal(1))
			requestedInstallationId := fakeService.GetInstallationLogsArgsForCall(0)
			Expect(requestedInstallationId).To(Equal(999))

			Expect(logger.PrintCallCount()).To(Equal(1))
			outputLogs := logger.PrintArgsForCall(0)[0]
			Expect(outputLogs).To(Equal("some log output"))
		})

		Context("Failure cases", func() {
			When("an unknown flag is provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{"--since", "yesterday"})
					Expect(err).To(MatchError("could not parse installation-log flags: flag provided but not defined: -since"))
				})
			})
			When("the installation id is not provided", func() {
				It("returns an error", func() {
					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not parse installation-log flags: missing required flag \"--id\""))
				})
			})
			When("the api fails to retrieve the installation log", func() {
				It("returns an error", func() {
					fakeService.GetInstallationLogsReturns(
						api.InstallationsServiceOutput{},
						errors.New("failed to retrieve installation log"),
					)
					err := command.Execute([]string{"--id", "999"})
					Expect(err).To(MatchError("failed to retrieve installation log"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewInstallationLog(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command retrieves the logs for a given installation.",
				ShortDescription: "output installation logs",
				Flags:            command.Options,
			}))
		})
	})
})
