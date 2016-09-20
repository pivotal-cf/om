package commands

import (
	"html/template"
	"io"
)

const usage = `om cli helps you interact with an OpsManager

Usage: om [options] <command> [<args>]
{{range .flags}}  {{.}}
{{end}}
Commands:
{{range .commands}}  {{.}}
{{end}}
`

type Helper interface {
	Help() string
}

type Help struct {
	flags    []Helper
	commands []Helper
	output   io.Writer
}

func NewHelp(flags []Helper, commands []Helper, output io.Writer) Help {
	return Help{
		flags:    flags,
		commands: commands,
		output:   output,
	}
}

func (h Help) Help() string {
	return "help     prints this usage information"
}

func (h Help) Execute() error {
	var flags []string
	for _, flag := range h.flags {
		flags = append(flags, flag.Help())
	}

	var commands []string
	for _, command := range append(h.commands, h) {
		commands = append(commands, command.Help())
	}

	t := template.Must(template.New("usage").Parse(usage))
	err := t.Execute(h.output, map[string]interface{}{
		"flags":    flags,
		"commands": commands,
	})
	if err != nil {
		return err
	}

	return nil
}
