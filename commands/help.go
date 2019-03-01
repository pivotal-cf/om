package commands

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/pivotal-cf/jhanda"
)

const tmpl = `{{.Title}}
{{.Description}}

Usage: {{.Usage}}
{{range .GlobalFlags}}  {{.}}
{{end}}
{{if .Arguments}}{{.ArgumentsName}}:
{{range .Arguments}}  {{.}}
{{end}}{{end}}
`

type TemplateContext struct {
	Title         string
	Description   string
	Usage         string
	GlobalFlags   []string
	ArgumentsName string
	Arguments     []string
}

type Help struct {
	output   io.Writer
	flags    string
	commands jhanda.CommandSet
}

func NewHelp(output io.Writer, flags string, commands jhanda.CommandSet) Help {
	return Help{
		output:   output,
		flags:    flags,
		commands: commands,
	}
}

func (h Help) Execute(args []string) error {
	globalFlags := strings.Split(h.flags, "\n")

	var context TemplateContext
	if len(args) == 0 {
		context = h.buildGlobalContext()
	} else {
		var err error
		context, err = h.buildCommandContext(args[0])
		if err != nil {
			return err
		}
	}

	context.GlobalFlags = globalFlags

	t := template.Must(template.New("usage").Parse(tmpl))

	err := t.Execute(h.output, context)
	if err != nil {
		return err
	}

	return nil
}

func (h Help) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command prints helpful usage information.",
		ShortDescription: "prints this usage information",
	}
}

func (h Help) buildGlobalContext() TemplateContext {
	var (
		length int
		names  []string
	)

	for name := range h.commands {
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
		commands = append(commands, fmt.Sprintf("%s  %s", name, command.Usage().ShortDescription))
	}

	return TemplateContext{
		Title:         "ॐ",
		Description:   "om helps you interact with an Ops Manager",
		Usage:         "om [options] <command> [<args>]",
		ArgumentsName: "Commands",
		Arguments:     commands,
	}
}

func (h Help) buildCommandContext(command string) (TemplateContext, error) {
	usage, err := h.commands.Usage(command)
	if err != nil {
		return TemplateContext{}, err
	}

	var (
		flagList        []string
		argsPlaceholder string
	)
	if usage.Flags != nil {
		flagUsage, err := jhanda.PrintUsage(usage.Flags)
		if err != nil {
			return TemplateContext{}, err
		}

		for _, flag := range strings.Split(flagUsage, "\n") {
			if len(flag) != 0 {
				flagList = append(flagList, flag)
			}
		}

		if len(flagList) != 0 {
			argsPlaceholder = " [<args>]"
		}
	}

	return TemplateContext{
		Title:         fmt.Sprintf("ॐ  %s", command),
		Description:   usage.Description,
		Usage:         fmt.Sprintf("om [options] %s%s", command, argsPlaceholder),
		ArgumentsName: "Command Arguments",
		Arguments:     flagList,
	}, nil
}

func (h Help) pad(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}
