package commands_test

import (
	"bytes"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	GLOBAL_USAGE = `ॐ
om helps you interact with an Ops Manager

Usage: om [options] <command> [<args>]
  --query, -?     asks a question
  --surprise, -!  gives you a present

Commands:
  bake   bakes you a cake
  clean  cleans up after baking
`

	COMMAND_USAGE = `ॐ  bake
This command will help you bake a cake.

Usage: om [options] bake [<args>]
  --query, -?     asks a question
  --surprise, -!  gives you a present

Command Arguments:
  --butter, -b  int (variadic)  sticks of butter
  --flour, -f   int             cups of flour
  --lemon, -l   int             teaspoons of lemon juice
`

	FLAGLESS_USAGE = `ॐ  bake
This command will help you bake a cake.

Usage: om [options] bake
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
		When("no command name is given", func() {
			It("prints the global usage to the output", func() {
				bake := &fakeCommand{
					usage: jhanda.Usage{ShortDescription: "bakes you a cake"},
				}

				clean := &fakeCommand{
					usage: jhanda.Usage{ShortDescription: "cleans up after baking"},
				}

				help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{
					"bake":  bake,
					"clean": clean,
				})
				err := help.Execute([]string{})
				Expect(err).NotTo(HaveOccurred())

				Expect(output.String()).To(ContainSubstring(GLOBAL_USAGE))
			})
		})

		When("a command name is given", func() {
			It("prints the usage for that command", func() {
				bake := &fakeCommand{
					usage: jhanda.Usage{
						Description:      "This command will help you bake a cake.",
						ShortDescription: "bakes you a cake",
						Flags: struct {
							Flour  int   `short:"f" long:"flour"  description:"cups of flour"`
							Butter []int `short:"b" long:"butter" description:"sticks of butter"`
							Lemon  int   `short:"l" long:"lemon"  description:"teaspoons of lemon juice"`
						}{},
					},
				}

				help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"bake": bake})
				err := help.Execute([]string{"bake"})
				Expect(err).NotTo(HaveOccurred())

				Expect(output.String()).To(ContainSubstring(COMMAND_USAGE))
			})

			When("the command does not exist", func() {
				It("returns an error", func() {
					help := commands.NewHelp(output, flags, jhanda.CommandSet{})
					err := help.Execute([]string{"missing-command"})
					Expect(err).To(MatchError("unknown command: missing-command"))
				})
			})

			When("the command flags cannot be determined", func() {
				It("returns an error", func() {
					bake := &fakeCommand{
						usage: jhanda.Usage{
							Description:      "This command will help you bake a cake.",
							ShortDescription: "bakes you a cake",
							Flags:            func() {},
						},
					}

					help := commands.NewHelp(output, flags, jhanda.CommandSet{"bake": bake})
					err := help.Execute([]string{"bake"})
					Expect(err).To(MatchError("unexpected pointer to non-struct type func"))
				})
			})

			When("there are no flags", func() {
				It("prints the usage of a flag-less command", func() {
					bake := &fakeCommand{
						usage: jhanda.Usage{
							Description:      "This command will help you bake a cake.",
							ShortDescription: "bakes you a cake",
						},
					}

					help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"bake": bake})
					err := help.Execute([]string{"bake"})
					Expect(err).NotTo(HaveOccurred())

					Expect(output.String()).To(ContainSubstring(FLAGLESS_USAGE))
					Expect(output.String()).NotTo(ContainSubstring("Command Arguments"))
				})
			})

			When("there is an empty flag object", func() {
				It("prints the usage of a flag-less command", func() {
					bake := &fakeCommand{
						usage: jhanda.Usage{
							Description:      "This command will help you bake a cake.",
							ShortDescription: "bakes you a cake",
							Flags:            struct{}{},
						},
					}

					help := commands.NewHelp(output, strings.TrimSpace(flags), jhanda.CommandSet{"bake": bake})
					err := help.Execute([]string{"bake"})
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

type fakeCommand struct {
	usage jhanda.Usage
}

func (f fakeCommand) Execute(args []string) error {
	return nil
}

func (f fakeCommand) Usage() jhanda.Usage {
	return f.usage
}
