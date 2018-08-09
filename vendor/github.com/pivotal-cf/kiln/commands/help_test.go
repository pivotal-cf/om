package commands_test

import (
	"bytes"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/kiln/commands"
	"github.com/pivotal-cf/kiln/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	GLOBAL_USAGE = `kiln
kiln helps you build ops manager compatible tiles

Usage: kiln [options] <command> [<args>]
  --query, -?     asks a question
  --surprise, -!  gives you a present

Commands:
  clean  cleans the pot you used
  cook   cooks you a stew
`

	COMMAND_USAGE = `kiln cook
This command will help you cook a stew.

Usage: kiln [options] cook [<args>]
  --query, -?     asks a question
  --surprise, -!  gives you a present

Command Arguments:
  --lemon, -l  int  teaspoons of lemon juice
  --stock, -s  int  teaspoons of vegan stock
  --water, -w  int  cups of water
`

	FLAGLESS_USAGE = `kiln cook
This command will help you cook a stew.

Usage: kiln [options] cook
  --query, -?     asks a question
  --surprise, -!  gives you a present
`
)

var _ = Describe("Help", func() {
	var (
		output *bytes.Buffer
		flags  string
	)

	BeforeEach(func() {
		output = bytes.NewBuffer([]byte{})
		flags = strings.TrimSpace(`
--query, -?     asks a question
--surprise, -!  gives you a present
`)
	})

	Describe("Execute", func() {
		Context("when no command name is given", func() {
			It("prints the global usage to the output", func() {
				cook := &fakes.Command{}
				cook.UsageReturns(jhanda.Usage{ShortDescription: "cooks you a stew"})

				clean := &fakes.Command{}
				clean.UsageReturns(jhanda.Usage{ShortDescription: "cleans the pot you used"})

				help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{
					"cook":  cook,
					"clean": clean,
				})
				err := help.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(output.String()).To(ContainSubstring(GLOBAL_USAGE))
			})
		})

		Context("when a command name is given", func() {
			It("prints the usage for that command", func() {
				cook := &fakes.Command{}
				cook.UsageReturns(jhanda.Usage{
					Description:      "This command will help you cook a stew.",
					ShortDescription: "cooks you a stew",
					Flags: struct {
						Water int `short:"w" long:"water"  description:"cups of water"`
						Stock int `short:"s" long:"stock"  description:"teaspoons of vegan stock"`
						Lemon int `short:"l" long:"lemon"  description:"teaspoons of lemon juice"`
					}{},
				})

				help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"cook": cook})
				err := help.Execute([]string{"cook"})
				Expect(err).NotTo(HaveOccurred())

				Expect(output.String()).To(ContainSubstring(COMMAND_USAGE))
			})

			Context("when the command does not exist", func() {
				It("returns an error", func() {
					help := commands.NewHelp(output, flags, jhanda.CommandSet{})
					err := help.Execute([]string{"missing-command"})
					Expect(err).To(MatchError("unknown command: missing-command"))
				})
			})

			Context("when the command flags cannot be determined", func() {
				It("returns an error", func() {
					cook := &fakes.Command{}
					cook.UsageReturns(jhanda.Usage{
						Description:      "This command will help you cook a stew.",
						ShortDescription: "cooks you a stew",
						Flags:            func() {},
					})

					help := commands.NewHelp(output, flags, jhanda.CommandSet{"cook": cook})
					err := help.Execute([]string{"cook"})
					Expect(err).To(MatchError("unexpected pointer to non-struct type func"))
				})
			})

			Context("when there are no flags", func() {
				It("prints the usage of a flag-less command", func() {
					cook := &fakes.Command{}
					cook.UsageReturns(jhanda.Usage{
						Description:      "This command will help you cook a stew.",
						ShortDescription: "cooks you a stew",
					})

					help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"cook": cook})
					err := help.Execute([]string{"cook"})
					Expect(err).NotTo(HaveOccurred())

					Expect(output.String()).To(ContainSubstring(FLAGLESS_USAGE))
					Expect(output.String()).NotTo(ContainSubstring("Command Arguments"))
				})
			})

			Context("when there is an empty flag object", func() {
				It("prints the usage of a flag-less command", func() {
					cook := &fakes.Command{}
					cook.UsageReturns(jhanda.Usage{
						Description:      "This command will help you cook a stew.",
						ShortDescription: "cooks you a stew",
						Flags:            struct{}{},
					})

					help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"cook": cook})
					err := help.Execute([]string{"cook"})
					Expect(err).NotTo(HaveOccurred())

					Expect(output.String()).To(ContainSubstring(FLAGLESS_USAGE))
					Expect(output.String()).NotTo(ContainSubstring("Command Arguments"))
				})
			})
		})
	})

	Describe("Usage", func() {
		It("returns usage information for the command", func() {
			help := commands.NewHelp(nil, "", jhanda.CommandSet{})
			Expect(help.Usage()).To(Equal(jhanda.Usage{
				Description:      "This command prints helpful usage information.",
				ShortDescription: "prints this usage information",
			}))
		})
	})
})
