package commands_test

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
)

var _ = Describe("DeleteInstallation", func() {
	var (
		service     *fakes.DeleteInstallationService
		logger      *fakes.Logger
		logsOutputs []api.InstallationsServiceOutput
		logsErrors  []error
		statusCount int
		logsCount   int
	)

	BeforeEach(func() {
		service = &fakes.DeleteInstallationService{}
		logger = &fakes.Logger{}

		statusCount = 0
		logsCount = 0

		service.GetCurrentInstallationLogsStub = func() (api.InstallationsServiceOutput, error) {
			output := logsOutputs[logsCount]
			err := logsErrors[logsCount]
			logsCount++
			return output, err
		}
	})

	Describe("Execute", func() {
		It("deletes the current installation in the Ops Manager", func() {
			service.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{ID: 311}
				return output, nil
			}

			logChan := make(chan string)
			errChan := make(chan error)
			logsOutputs = []api.InstallationsServiceOutput{
				{LogChan: logChan, ErrorChan: errChan},
			}

			logsErrors = []error{nil}

			go func() {
				logChan <- "some thing"
				close(logChan)
				close(errChan)
			}()

			logsErrors = []error{nil, nil, nil}

			command := commands.NewDeleteInstallation(service, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))

			Expect(service.DeleteInstallationAssetCollectionCallCount()).To(Equal(1))

			content = logger.PrintlnArgsForCall(0)
			Expect(fmt.Sprint(content...)).To(Equal("some thing"))
		})

		It("handles a failed installation", func() {
			service.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{ID: 311}
				return output, nil
			}

			logChan := make(chan string)
			errChan := make(chan error)
			logsOutputs = []api.InstallationsServiceOutput{
				{LogChan: logChan, ErrorChan: errChan},
			}

			logsErrors = []error{nil}

			go func() {
				logChan <- "some thing"
				close(logChan)
				errChan <- api.InstallFailed
				close(errChan)
			}()

			command := commands.NewDeleteInstallation(service, logger)

			err := command.Execute([]string{})
			Expect(err).To(MatchError("deleting the installation was unsuccessful"))
		})

		It("handles the case when there is no installation to delete", func() {
			service.DeleteInstallationAssetCollectionStub = func() (api.InstallationsServiceOutput, error) {
				output := api.InstallationsServiceOutput{}
				return output, nil
			}

			command := commands.NewDeleteInstallation(service, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to delete the installation on the targeted Ops Manager"))
			format, content = logger.PrintfArgsForCall(1)
			Expect(fmt.Sprintf(format, content...)).To(Equal("no installation to delete"))
		})

		Context("when an installation is already running", func() {
			It("re-attaches to the installation", func() {
				service.RunningInstallationReturns(api.InstallationsServiceOutput{ID: 311, Status: "running"}, nil)

				logChan := make(chan string)
				errChan := make(chan error)
				logsOutputs = []api.InstallationsServiceOutput{
					{LogChan: logChan, ErrorChan: errChan},
				}

				logsErrors = []error{nil}

				go func() {
					logChan <- "some thing"
					close(logChan)
					close(errChan)
				}()

				command := commands.NewDeleteInstallation(service, logger)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.DeleteInstallationAssetCollectionCallCount()).To(Equal(0))

				format, content := logger.PrintfArgsForCall(0)
				Expect(fmt.Sprintf(format, content...)).To(Equal("found already running deletion...attempting to re-attach"))
			})
		})

		Context("failure cases", func() {
			Context("when the delete to installation_asset_collection is unsuccessful", func() {
				It("returns an error", func() {
					service.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewDeleteInstallation(service, logger)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("failed to delete installation: some error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					service.DeleteInstallationAssetCollectionReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					logChan := make(chan string)
					errChan := make(chan error)
					logsOutputs = []api.InstallationsServiceOutput{
						{LogChan: logChan, ErrorChan: errChan},
					}

					logsErrors = []error{errors.New("no")}

					command := commands.NewDeleteInstallation(service, logger)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get logs: no"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewDeleteInstallation(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command deletes all the products installed on the targeted Ops Manager.",
				ShortDescription: "deletes all the products on the Ops Manager targeted",
			}))
		})
	})
})
