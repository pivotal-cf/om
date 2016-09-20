package commands_test

import (
	"bytes"

	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const USAGE = `om cli helps you interact with an OpsManager

Usage: om [options] <command> [<args>]
  -?, --query     asks a question
  -!, --surprise  gives you a present

Commands:
  bake     bakes you a cake
  clean    cleans up after baking
  help     prints this usage information
`

var _ = Describe("Help", func() {
	Describe("Execute", func() {
		It("prints the global usage to the output", func() {
			output := bytes.NewBuffer([]byte{})

			query := &fakes.Helper{}
			query.HelpCall.Returns.Help = "-?, --query     asks a question"

			surprise := &fakes.Helper{}
			surprise.HelpCall.Returns.Help = "-!, --surprise  gives you a present"

			bake := &fakes.Helper{}
			bake.HelpCall.Returns.Help = "bake     bakes you a cake"

			clean := &fakes.Helper{}
			clean.HelpCall.Returns.Help = "clean    cleans up after baking"

			help := commands.NewHelp([]commands.Helper{query, surprise}, []commands.Helper{bake, clean}, output)
			err := help.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(output.String()).To(ContainSubstring(USAGE))
		})
	})

	Describe("Help", func() {
		It("returns a short help description of the command", func() {
			help := commands.NewHelp(nil, nil, nil)
			Expect(help.Help()).To(Equal("help     prints this usage information"))
		})
	})
})
