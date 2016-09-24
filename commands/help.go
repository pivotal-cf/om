package commands

import (
	"fmt"
	"html/template"
	"io"
	"sort"
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

type Help struct {
	output   io.Writer
	flags    string
	commands Set
}

func NewHelp(output io.Writer, flags string, commands Set) Help {
	return Help{
		output:   output,
		flags:    flags,
		commands: commands,
	}
}

func (h Help) Help() string {
	return "prints this usage information"
}

func (h Help) Execute([]string) error {
	var flags []string
	for _, flag := range strings.Split(h.flags, "\n") {
		flags = append(flags, flag)
	}

	var (
		length int
		names  []string
	)

	for name, _ := range h.commands {
		names = append(names, name)
		if len(name) > length {
			length = len(name)
		}
	}

	sort.Strings(names)

	var commands []string
	for _, name := range names {
		command := h.commands[name]
		name = h.pad(name, " ", length)
		commands = append(commands, fmt.Sprintf("%s  %s", name, command.Help()))
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

func (h Help) pad(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}
