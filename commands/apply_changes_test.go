package commands_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type netError struct {
	error
}

func (ne netError) Temporary() bool {
	return true
}

func (ne netError) Timeout() bool {
	return false
}

var _ = Describe("ApplyChanges", func() {
	var (
		service       *fakes.ApplyChangesService
		logger        *fakes.Logger
		writer        *fakes.LogWriter
		statusOutputs []api.InstallationsServiceOutput
		statusErrors  []error
		logsOutputs   []api.InstallationsServiceOutput
		logsErrors    []error
		statusCount   int
		logsCount     int
	)

	BeforeEach(func() {
		service = &fakes.ApplyChangesService{}
		logger = &fakes.Logger{}
		writer = &fakes.LogWriter{}

		statusCount = 0
		logsCount = 0

		service.GetInstallationStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := statusOutputs[statusCount]
			err := statusErrors[statusCount]
			statusCount++
			return output, err
		}

		service.GetInstallationLogsStub = func(id int) (api.InstallationsServiceOutput, error) {
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

			command := commands.NewApplyChanges(service, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(1))

			ignoreWarnings, deployProducts, _ := service.CreateInstallationArgsForCall(0)
			Expect(ignoreWarnings).To(Equal(false))
			Expect(deployProducts).To(Equal(true))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to apply changes to the targeted Ops Manager"))

			Expect(service.GetInstallationArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationCallCount()).To(Equal(3))

			Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(311))
			Expect(service.GetInstallationLogsCallCount()).To(Equal(3))

			Expect(writer.FlushCallCount()).To(Equal(3))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
			Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
		})

		Context("when passed the ignore-warnings flag", func() {
			It("applies changes while ignoring warnings", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

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
				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{"--ignore-warnings"})
				Expect(err).NotTo(HaveOccurred())

				ignoreWarnings, _, _ := service.CreateInstallationArgsForCall(0)
				Expect(ignoreWarnings).To(Equal(true))
			})
		})

		Context("when passed the skip-deploy-products flag", func() {
			It("applies changes while not deploying products", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

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
				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{"--skip-deploy-products"})
				Expect(err).NotTo(HaveOccurred())

				_, deployProducts, _ := service.CreateInstallationArgsForCall(0)
				Expect(deployProducts).To(Equal(false))
			})

			It("fails if product names were specified", func() {
				command := commands.NewApplyChanges(service, writer, logger, 1)
				err := command.Execute([]string{"--skip-deploy-products", "--product-name", "product1"})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when passed the product-name flag", func() {
			It("passes product names to the installation service", func() {
				service.CreateInstallationReturns(api.InstallationsServiceOutput{}, errors.New("error"))
				service.RunningInstallationReturns(api.InstallationsServiceOutput{}, nil)

				command := commands.NewApplyChanges(service, writer, logger, 1)
				err := command.Execute([]string{"--product-name", "product1", "--product-name", "product2"})
				Expect(err).To(HaveOccurred())

				_, _, productNames := service.CreateInstallationArgsForCall(0)
				Expect(productNames).To(ConsistOf("product1", "product2"))
			})
		})

		It("re-attaches to an ongoing installation", func() {
			installationStartedAt := time.Date(2017, time.February, 25, 02, 31, 1, 0, time.UTC)

			service.RunningInstallationReturns(api.InstallationsServiceOutput{
				ID:        200,
				Status:    "running",
				StartedAt: &installationStartedAt,
			}, nil)

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

			command := commands.NewApplyChanges(service, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			Expect(service.CreateInstallationCallCount()).To(Equal(0))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("found already running installation...re-attaching (Installation ID: 200, Started: Sat Feb 25 02:31:01 UTC 2017)"))

			Expect(service.GetInstallationArgsForCall(0)).To(Equal(200))
			Expect(service.GetInstallationLogsArgsForCall(0)).To(Equal(200))
		})

		It("handles a failed installation", func() {
			service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)
			statusOutputs = []api.InstallationsServiceOutput{
				{Status: "failed"},
			}

			statusErrors = []error{nil}

			logsOutputs = []api.InstallationsServiceOutput{
				{Logs: "start of logs"},
			}

			logsErrors = []error{nil}

			command := commands.NewApplyChanges(service, writer, logger, 1)

			err := command.Execute([]string{})
			Expect(err).To(MatchError("installation was unsuccessful"))
		})

		Context("failure cases", func() {
			Context("when checking for an already running installation returns an error", func() {
				It("returns an error", func() {
					service.RunningInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("could not check for any already running installation: some error"))
				})
			})

			Context("when an installation cannot be triggered", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to trigger: some error"))
				})
			})

			Context("when getting the installation status has an error", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{{}}

					statusErrors = []error{errors.New("another error")}

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get status: another error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{}}

					logsErrors = []error{errors.New("no")}

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get logs: no"))
				})
			})

			Context("when there is an error flushing the logs", func() {
				It("returns an error", func() {
					service.CreateInstallationReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{
						{Status: "running"},
					}

					statusErrors = []error{nil}

					logsOutputs = []api.InstallationsServiceOutput{{Logs: "some logs"}}

					logsErrors = []error{nil}

					writer.FlushReturns(errors.New("yes"))

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to flush logs: yes"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			command := commands.NewApplyChanges(nil, nil, nil, 1)
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
				ShortDescription: "triggers an install on the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
