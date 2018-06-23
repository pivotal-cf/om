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
	"time"
)

var _ = Describe("ApplyChanges", func() {
	var (
		service     *fakes.ApplyChangesService
		logger      *fakes.Logger
		writer      *fakes.LogWriter
		logsOutputs []api.InstallationsServiceOutput
		logsErrors  []error
		statusCount int
		logsCount   int
	)

	BeforeEach(func() {
		service = &fakes.ApplyChangesService{}
		logger = &fakes.Logger{}
		writer = &fakes.LogWriter{}

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
		It("applies changes to the Ops Manager", func() {
			service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
			service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

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

			command := commands.NewApplyChanges(service, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(1))

			ignoreWarnings, deployProducts := service.CreateInstallationArgsForCall(0)
			Expect(ignoreWarnings).To(Equal(false))
			Expect(deployProducts).To(Equal(true))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to apply changes to the targeted Ops Manager"))

			Expect(service.GetCurrentInstallationLogsCallCount()).To(Equal(1))

			content = logger.PrintlnArgsForCall(0)
			Expect(fmt.Sprint(content...)).To(Equal("some thing"))
		})

		Context("when passed the ignore-warnings flag", func() {
			It("applies changes while ignoring warnings", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

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

				command := commands.NewApplyChanges(service, logger)

				err := command.Execute([]string{"--ignore-warnings"})
				Expect(err).NotTo(HaveOccurred())

				ignoreWarnings, _ := service.CreateInstallationArgsForCall(0)
				Expect(ignoreWarnings).To(Equal(true))
			})
		})

		Context("when passed the skip-deploy-products flag", func() {
			It("applies changes while not deploying products", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

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

				command := commands.NewApplyChanges(service, logger)

				err := command.Execute([]string{"--skip-deploy-products"})
				Expect(err).NotTo(HaveOccurred())

				_, deployProducts := service.CreateInstallationArgsForCall(0)
				Expect(deployProducts).To(Equal(false))
			})
		})

		It("re-attaches to an ongoing installation", func() {
			installationStartedAt := time.Date(2017, time.February, 25, 02, 31, 1, 0, time.UTC)

			service.RunningInstallationReturns(api.InstallationsServiceOutput{
				ID:        200,
				Status:    "running",
				StartedAt: &installationStartedAt,
			}, nil)

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

			command := commands.NewApplyChanges(service, logger)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(0))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("found already running installation...re-attaching (Installation ID: 200, Started: Sat Feb 25 02:31:01 UTC 2017)"))
		})

		It("handles a failed installation", func() {
			service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)

			logChan := make(chan string)
			errChan := make(chan error)
			logsOutputs = []api.InstallationsServiceOutput{
				{LogChan: logChan, ErrorChan: errChan},
			}

			logsErrors = []error{nil}

			go func() {
				logChan <- "some thing"
				close(logChan)
				errChan <- errors.New("installation was unsuccessful")
				close(errChan)
			}()

			command := commands.NewApplyChanges(service, logger)

			err := command.Execute([]string{})
			Expect(err).To(MatchError("installation was unsuccessful"))
		})

		Context("failure cases", func() {
			Context("when checking for an already running installation returns an error", func() {
				It("returns an error", func() {
					service.RunningInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewApplyChanges(service, logger)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not check for any already running installation: some error"))
				})
			})

			Context("when an installation cannot be triggered", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewApplyChanges(service, logger)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to trigger: some error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					logChan := make(chan string)
					errChan := make(chan error)
					logsOutputs = []api.InstallationsServiceOutput{
						{LogChan: logChan, ErrorChan: errChan},
					}

					logsErrors = []error{errors.New("no")}

					command := commands.NewApplyChanges(service, logger)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get logs: no"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewApplyChanges(nil, nil)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
				ShortDescription: "triggers an install on the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
