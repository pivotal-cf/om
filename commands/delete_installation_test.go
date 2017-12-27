package commands_test

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteInstallation", func() {
	var (
		installationService *fakes.InstallationsService
		deleteService       *fakes.InstallationAssetDeleterService
		logger              *fakes.Logger
		writer              *fakes.LogWriter
		statusOutputs       []api.InstallationsServiceOutput
		statusErrors        []error
		logsOutputs         []api.InstallationsServiceOutput
		logsErrors          []error
		statusCount         int
		logsCount           int
	)

	BeforeEach(func() {
		installationService = &fakes.InstallationsService{}
		deleteService = &fakes.InstallationAssetDeleterService{}
		logger = &fakes.Logger{}
		writer = &fakes.LogWriter{}

		statusCount = 0
		logsCount = 0

		installationService.StatusStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := statusOutputs[statusCount]
			err := statusErrors[statusCount]
			statusCount++
			return output, err
		}

		installationService.LogsStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := logsOutputs[logsCount]
			err := logsErrors[logsCount]
			logsCount++
			return output, err
		}

	})

	Describe("Execute", func() {
		It("deletes the current installation in the Ops Manager", func() {
			deleteService.DeleteStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{ID: 311}
				return output, nil
			}

			statusOutputs = []api.InstallationsServiceOutput{
				{Status: "running"},
				{Status: "running"},
				{Status: "succeeded"},
			}

			statusErrors = []error{nil, nil, nil}

			logsOutputs = []api.InstallationsServiceOutput{
				{Logs: "start of logs"},
				{Logs: "these logs"},
				{Logs: "some other logs"},
			}

			logsErrors = []error{nil, nil, nil}

			command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))

			Expect(deleteService.DeleteCallCount()).To(Equal(1))
			Expect(installationService.StatusArgsForCall(0)).To(Equal(311))
			Expect(installationService.StatusCallCount()).To(Equal(3))

			Expect(installationService.LogsArgsForCall(0)).To(Equal(311))
			Expect(installationService.LogsCallCount()).To(Equal(3))

			Expect(writer.FlushCallCount()).To(Equal(3))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
			Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
		})

		It("handles a failed installation", func() {
			deleteService.DeleteStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{ID: 311}
				return output, nil
			}

			statusOutputs = []api.InstallationsServiceOutput{
				{Status: "failed"},
			}

			statusErrors = []error{nil}

			logsOutputs = []api.InstallationsServiceOutput{
				{Logs: "start of logs"},
			}

			logsErrors = []error{nil}

			command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).To(MatchError("deleting the installation was unsuccessful"))
		})

		It("handles the case when there is no installation to delete", func() {
			deleteService.DeleteStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{}
				return output, nil
			}

			command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))
			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("no installation to delete"))
		})

		Context("when an installation is already running", func() {
			It("re-attaches to the installation", func() {
				installationService.RunningInstallationReturns(api.InstallationsServiceOutput{ID: 311, Status: "running"}, nil)

				statusOutputs = []api.InstallationsServiceOutput{
					{Status: "running"},
					{Status: "running"},
					{Status: "succeeded"},
				}

				statusErrors = []error{nil, nil, nil}

				logsOutputs = []api.InstallationsServiceOutput{
					{Logs: "start of logs"},
					{Logs: "these logs"},
					{Logs: "some other logs"},
				}

				logsErrors = []error{nil, nil, nil}

				command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(deleteService.DeleteCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("found already running deletion...attempting to re-attach"))

				Expect(installationService.StatusArgsForCall(0)).To(Equal(311))
				Expect(installationService.LogsArgsForCall(0)).To(Equal(311))
			})
		})

		Context("failure cases", func() {
			Context("when the delete to installation_asset_collection is unsuccessful", func() {
				It("returns an error", func() {
					deleteService.DeleteReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to delete installation: some error"))
				})
			})

			Context("when getting the installation status has an error", func() {
				It("returns an error", func() {
					deleteService.DeleteReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{{}}

					statusErrors = []error{errors.New("another error")}

					command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get status: another error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					deleteService.DeleteReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{}}

					logsErrors = []error{errors.New("no")}

					command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get logs: no"))
				})
			})

			Context("when there is an error flushing the logs", func() {
				It("returns an error", func() {
					deleteService.DeleteReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{Logs: "some logs"}}

					logsErrors = []error{nil}

					writer.FlushReturns(errors.New("yes"))

					command := commands.NewDeleteInstallation(deleteService, installationService, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to flush logs: yes"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDeleteInstallation(nil, nil, nil, nil, 1)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command deletes all the products installed on the targeted Ops Manager.",
				ShortDescription: "deletes all the products on the Ops Manager targeted",
			}))
		})
	})
})
