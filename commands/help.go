package commands

import (
	"html/template"
	"io"
	"strings"
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
	output   io.Writer
	flags    string
	commands []Helper
}

func NewHelp(output io.Writer, flags string, commands ...Helper) Help {
	return Help{
		output:   output,
		flags:    flags,
		commands: commands,
	}
}

func (h Help) Execute([]string) error {
	var flags []string
	for _, flag := range strings.Split(h.flags, "\n") {
		flags = append(flags, flag)
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

func (h Help) Help() string {
	return "help     prints this usage information"
}
