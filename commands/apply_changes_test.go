package commands_test

import (
	"errors"
	"fmt"
	"io"

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
		service       *fakes.InstallationsService
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
		service = &fakes.InstallationsService{}
		logger = &fakes.Logger{}
		writer = &fakes.LogWriter{}

		statusCount = 0
		logsCount = 0

		service.StatusStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := statusOutputs[statusCount]
			err := statusErrors[statusCount]
			statusCount++
			return output, err
		}

		service.LogsStub = func(id int) (api.InstallationsServiceOutput, error) {
			output := logsOutputs[logsCount]
			err := logsErrors[logsCount]
			logsCount++
			return output, err
		}

	})

	Describe("Execute", func() {
		It("applies changes to the Ops Manager", func() {
			service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
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

			Expect(service.TriggerCallCount()).To(Equal(1))
			Expect(service.TriggerArgsForCall(0)).To(Equal(false))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("attempting to apply changes to the targeted Ops Manager"))

			Expect(service.StatusArgsForCall(0)).To(Equal(311))
			Expect(service.StatusCallCount()).To(Equal(3))

			Expect(service.LogsArgsForCall(0)).To(Equal(311))
			Expect(service.LogsCallCount()).To(Equal(3))

			Expect(writer.FlushCallCount()).To(Equal(3))
			Expect(writer.FlushArgsForCall(0)).To(Equal("start of logs"))
			Expect(writer.FlushArgsForCall(1)).To(Equal("these logs"))
			Expect(writer.FlushArgsForCall(2)).To(Equal("some other logs"))
		})

		Context("when passed the ignore-warnings flag", func() {
			It("applies changes while ignoring warnings", func() {
				service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
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

				Expect(service.TriggerArgsForCall(0)).To(Equal(true))
			})
		})

		It("re-attaches to an ongoing installation", func() {
			service.RunningInstallationReturns(api.InstallationsServiceOutput{ID: 200, Status: "running"}, nil)

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

			Expect(service.TriggerCallCount()).To(Equal(0))

			format, content := logger.PrintfArgsForCall(0)
			Expect(fmt.Sprintf(format, content...)).To(Equal("found already running installation...re-attaching"))

			Expect(service.StatusArgsForCall(0)).To(Equal(200))
			Expect(service.LogsArgsForCall(0)).To(Equal(200))
		})

		It("handles a failed installation", func() {
			service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
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

		Context("when there are network errors during status check", func() {
			It("ignores the temporary network error", func() {
				service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				statusOutputs = []api.InstallationsServiceOutput{
					{},
					{Status: "running"},
					{Status: "succeeded"},
				}

				statusErrors = []error{netError{errors.New("whoops")}, nil, nil}

				logsOutputs = []api.InstallationsServiceOutput{
					{Logs: "something logged"},
					{Logs: "another thing logged"},
				}

				logsErrors = []error{nil, nil, nil}

				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.StatusCallCount()).To(Equal(3))
				Expect(service.LogsCallCount()).To(Equal(2))

				Expect(writer.FlushCallCount()).To(Equal(2))
				Expect(writer.FlushArgsForCall(0)).To(Equal("something logged"))
				Expect(writer.FlushArgsForCall(1)).To(Equal("another thing logged"))
			})
		})

		Context("when there are network errors during log fetching", func() {
			It("ignores the temporary network error", func() {
				service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				statusOutputs = []api.InstallationsServiceOutput{
					{Status: "running"},
					{Status: "running"},
					{Status: "succeeded"},
				}

				statusErrors = []error{nil, nil, nil}

				logsOutputs = []api.InstallationsServiceOutput{
					{Logs: "one log"},
					{},
					{Logs: "two log"},
				}

				logsErrors = []error{nil, netError{errors.New("no")}, nil}

				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.StatusCallCount()).To(Equal(3))
				Expect(service.LogsCallCount()).To(Equal(3))

				Expect(writer.FlushCallCount()).To(Equal(2))
				Expect(writer.FlushArgsForCall(0)).To(Equal("one log"))
				Expect(writer.FlushArgsForCall(1)).To(Equal("two log"))
			})
		})

		Context("when there are EOF errors during status check", func() {
			It("ignores the temporary EOF error", func() {
				service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				statusOutputs = []api.InstallationsServiceOutput{
					{Status: "running"},
					{},
					{Status: "succeeded"},
				}

				statusErrors = []error{nil, io.EOF, nil}

				logsOutputs = []api.InstallationsServiceOutput{
					{Logs: "one"},
					{Logs: "two"},
				}

				logsErrors = []error{nil, nil, nil}

				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.StatusCallCount()).To(Equal(3))
				Expect(service.LogsCallCount()).To(Equal(2))

				Expect(writer.FlushCallCount()).To(Equal(2))
				Expect(writer.FlushArgsForCall(0)).To(Equal("one"))
				Expect(writer.FlushArgsForCall(1)).To(Equal("two"))
			})
		})

		Context("when there are EOF errors during log fetching", func() {
			It("ignores the temporary EOF error", func() {
				service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)
				statusOutputs = []api.InstallationsServiceOutput{
					{Status: "running"},
					{Status: "running"},
					{Status: "running"},
					{Status: "succeeded"},
				}

				statusErrors = []error{nil, nil, nil, nil}

				logsOutputs = []api.InstallationsServiceOutput{
					{Logs: "one log"},
					{Logs: "two log"},
					{},
					{Logs: "three log"},
				}

				logsErrors = []error{nil, nil, io.EOF, nil}

				command := commands.NewApplyChanges(service, writer, logger, 1)

				err := command.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(service.StatusCallCount()).To(Equal(4))
				Expect(service.LogsCallCount()).To(Equal(4))

				Expect(writer.FlushCallCount()).To(Equal(3))
				Expect(writer.FlushArgsForCall(0)).To(Equal("one log"))
				Expect(writer.FlushArgsForCall(1)).To(Equal("two log"))
				Expect(writer.FlushArgsForCall(2)).To(Equal("three log"))
			})
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
					service.TriggerReturns(api.InstallationsServiceOutput{}, errors.New("some error"))

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to trigger: some error"))
				})
			})

			Context("when getting the installation status has an error", func() {
				It("returns an error", func() {
					service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)

					statusOutputs = []api.InstallationsServiceOutput{{}}

					statusErrors = []error{errors.New("another error")}

					command := commands.NewApplyChanges(service, writer, logger, 1)

					err := command.Execute([]string{})
					Expect(err).To(MatchError("installation failed to get status: another error"))
				})
			})

			Context("when there is an error fetching the logs", func() {
				It("returns an error", func() {
					service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)

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
					service.TriggerReturns(api.InstallationsServiceOutput{ID: 311}, nil)

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
			Expect(command.Usage()).To(Equal(commands.Usage{
				Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
				ShortDescription: "triggers an install on the Ops Manager targeted",
				Flags:            command.Options,
			}))
		})
	})
})
