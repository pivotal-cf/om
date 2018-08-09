package jhanda_test

import (
	"errors"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/jhanda/internal/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Set", func() {
	It("executes the given command", func() {
		command := &fakes.Command{}

		commandSet := jhanda.CommandSet{
			"my-command": command,
		}

		err := commandSet.Execute("my-command", []string{"--arg-1", "--arg-2"})
		Expect(err).NotTo(HaveOccurred())

		Expect(command.ExecuteArgsForCall(0)).To(Equal([]string{"--arg-1", "--arg-2"}))
	})

	Context("when the given command does not exist", func() {
		It("returns an error", func() {
			commandSet := jhanda.CommandSet{}

			err := commandSet.Execute("missing-command", []string{})
			Expect(err).To(MatchError("unknown command: missing-command"))
		})
	})

	Context("failure cases", func() {
		Context("when the command execution errors", func() {
			It("returns an error", func() {
				command := &fakes.Command{}
				command.ExecuteReturns(errors.New("failed to execute"))

				commandSet := jhanda.CommandSet{
					"erroring-command": command,
				}

				err := commandSet.Execute("erroring-command", []string{})
				Expect(err).To(MatchError("could not execute \"erroring-command\": failed to execute"))
			})
		})
	})

	Describe("when --help is passed as an argument", func() {
		It("executes the help for the command", func() {
			command := &fakes.Command{}
			helpCommand := &fakes.Command{}

			commandSet := jhanda.CommandSet{
				"my-command": command,
				"help":       helpCommand,
			}

			err := commandSet.Execute("my-command", []string{"--arg-1", "--help", "--arg-2"})
			Expect(err).NotTo(HaveOccurred())

			Expect(command.ExecuteCallCount()).To(Equal(0))
			Expect(helpCommand.ExecuteArgsForCall(0)).To(Equal([]string{"my-command"}))
		})
	})

	Describe("when -h is passed as an argument", func() {
		It("executes the help for the command", func() {
			command := &fakes.Command{}
			helpCommand := &fakes.Command{}

			commandSet := jhanda.CommandSet{
				"my-command": command,
				"help":       helpCommand,
			}

			err := commandSet.Execute("my-command", []string{"--arg-1", "-h", "--arg-2"})
			Expect(err).NotTo(HaveOccurred())

			Expect(command.ExecuteCallCount()).To(Equal(0))
			Expect(helpCommand.ExecuteArgsForCall(0)).To(Equal([]string{"my-command"}))
		})
	})

	Describe("when -help is passed as an argument", func() {
		It("executes the help for the command", func() {
			command := &fakes.Command{}
			helpCommand := &fakes.Command{}

			commandSet := jhanda.CommandSet{
				"my-command": command,
				"help":       helpCommand,
			}

			err := commandSet.Execute("my-command", []string{"--arg-1", "-help", "--arg-2"})
			Expect(err).NotTo(HaveOccurred())

			Expect(command.ExecuteCallCount()).To(Equal(0))
			Expect(helpCommand.ExecuteArgsForCall(0)).To(Equal([]string{"my-command"}))
		})
	})

	Describe("Usage", func() {
		It("returns the usage information for the given command", func() {
			command := &fakes.Command{}
			command.UsageReturns(jhanda.Usage{Description: "my-command description"})

			commandSet := jhanda.CommandSet{
				"my-command": command,
			}

			usage, err := commandSet.Usage("my-command")
			Expect(err).NotTo(HaveOccurred())

			Expect(usage).To(Equal(jhanda.Usage{Description: "my-command description"}))
		})

		Context("when the given command does not exist", func() {
			It("returns an error", func() {
				commandSet := jhanda.CommandSet{}

				_, err := commandSet.Usage("missing-command")
				Expect(err).To(MatchError("unknown command: missing-command"))
			})
		})
	})
})
