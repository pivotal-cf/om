package commands_test

import (
	"bytes"
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
		fakeService   *fakes.DeleteInstallationService
		logger        *fakes.Logger
		writer        *fakes.LogWriter
		statusOutputs []api.InstallationsServiceOutput
		statusErrors  []error
		logsOutputs   []api.InstallationsServiceOutput
		logsErrors    []error
		statusCount   int
		logsCount     int
		stdin         *bytes.Buffer
	)

	BeforeEach(func() {
		fakeService = &fakes.DeleteInstallationService{}
		logger = &fakes.Logger{}
		writer = &fakes.LogWriter{}
		stdin = bytes.NewBuffer([]byte{})

		statusCount = 0
		logsCount = 0

		fakeService.GetInstallationStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := statusOutputs[statusCount]
			err := statusErrors[statusCount]
			statusCount++
			return output, err
		}

		fakeService.GetInstallationLogsStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := logsOutputs[logsCount]
			err := logsErrors[logsCount]
			logsCount++
			return output, err
		}

	})

	Describe("Execute", func() {
		It("deletes the current installation in the Ops Manager", func() {
			fakeService.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
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

			command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

			_, err := stdin.WriteString("y\n")
			Expect(err).NotTo(HaveOccurred())

			err = command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("please press y to confirm deletion: "))

			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))

			Expect(fakeService.DeleteInstallationAssetCollectionCallCount()).To(Equal(1))
			Expect(fakeService.GetInstallationArgsForCall(0)).To(Equal(311))
			Expect(fakeService.GetInstallationCallCount()).To(Equal(3))

			Expect(fakeService.GetInstallationLogsArgsForCall(0)).To(Equal(311))
			Expect(fakeService.GetInstallationLogsCallCount()).To(Equal(3))

			Expect(writer.FlushCallCount()).To(Equal(3))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
			Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
		})

		It("handles a failed installation", func() {
			fakeService.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
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

			command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

			err := command.Execute([]string{"--force"})
			Expect(err).To(MatchError("deleting the installation was unsuccessful"))
		})

		It("handles the case when there is no installation to delete", func() {
			fakeService.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{}
				return output, nil
			}

			command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

			err := command.Execute([]string{"--force"})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))
			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("no installation to delete"))
		})

		Context("when an installation is already running", func() {
			It("re-attaches to the installation", func() {
				fakeService.RunningInstallationReturns(api.InstallationsServiceOutput{ID: 311, Status: "running"}, nil)

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

				command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

				err := command.Execute([]string{"--force"})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeService.DeleteInstallationAssetCollectionCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("found already running deletion...attempting to re-attach"))

				Expect(fakeService.GetInstallationArgsForCall(0)).To(Equal(311))
				Expect(fakeService.GetInstallationLogsArgsForCall(0)).To(Equal(311))
			})
		})

		Context("failure cases", func() {
			Context("when the delete to installation_asset_collection is unsuccessful", func() {
				It("returns an error", func() {
					fakeService.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

					err := command.Execute([]string{"--force"})
					Expect(err).To(MatchError("failed to delete installation: some error"))
				})
			})

			Context("when getting the installation status has an error", func() {
				It("returns an error", func() {
					fakeService.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{{}}

					statusErrors = []error{errors.New("another error")}

					command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

					err := command.Execute([]string{"--force"})
					Expect(err).To(MatchError("installation failed to get status: another error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					fakeService.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{}}

					logsErrors = []error{errors.New("no")}

					command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

					err := command.Execute([]string{"--force"})
					Expect(err).To(MatchError("installation failed to get logs: no"))
				})
			})

			Context("when there is an error flushing the logs", func() {
				It("returns an error", func() {
					fakeService.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{Logs: "some logs"}}

					logsErrors = []error{nil}

					writer.FlushReturns(errors.New("failed flush"))

					command := commands.NewDeleteInstallation(fakeService, writer, logger, stdin, 1)

					err := command.Execute([]string{"--force"})
					Expect(err).To(MatchError("installation failed to flush logs: failed flush"))
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
				Flags:            command.Options,
			}))
		})
	})
})
