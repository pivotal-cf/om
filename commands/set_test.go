package commands_test

import (
	"errors"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Set", func() {
	Describe("Execute", func() {
		It("executes the given command", func() {
			command := &fakes.Command{}

			commandSet := commands.Set{
				"my-command": command,
			}

			err := commandSet.Execute("my-command")
			Expect(err).NotTo(HaveOccurred())

			Expect(command.ExecuteCall.CallCount).To(Equal(1))
		})

		Context("when the given command does not exist", func() {
			It("returns an error", func() {
				commandSet := commands.Set{}

				err := commandSet.Execute("missing-command")
				Expect(err).To(MatchError("unknown command: missing-command"))
			})
		})

		Context("failure cases", func() {
			Context("when the command execution errors", func() {
				It("returns an error", func() {
					command := &fakes.Command{}
					command.ExecuteCall.Returns.Error = errors.New("failed to execute")

					commandSet := commands.Set{
						"erroring-command": command,
					}

					err := commandSet.Execute("erroring-command")
					Expect(err).To(MatchError("could not execute \"erroring-command\": failed to execute"))
				})
			})
		})
	})
})
