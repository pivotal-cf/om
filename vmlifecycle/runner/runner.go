package runner

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/onsi/gomega/gexec"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	command string
	out     io.Writer
	err     io.Writer
}

func NewRunner(command string, stdout io.Writer, stderr io.Writer) (*Runner, error) {
	_, err := exec.LookPath(command)
	if err != nil {
		return nil, fmt.Errorf("the cli '%s' is not available in PATH: %w", command, err)
	}

	return &Runner{
		command: command,
		out:     stdout,
		err:     stderr,
	}, nil
}

func (r *Runner) Execute(args []interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
	return r.ExecuteWithEnvVars([]string{}, args)
}

func (r *Runner) ExecuteWithEnvVars(env []string, args []interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
	var outBufWriter bytes.Buffer
	var errBufWriter bytes.Buffer

	outWriter := gexec.NewPrefixedWriter(r.command+"["+color.GreenString("stdout")+"]: ", r.out)
	errWriter := gexec.NewPrefixedWriter(r.command+"["+color.RedString("stderr")+"]: ", r.err)

	stringArgs := []string{}
	shownArgs := []string{}
	for _, arg := range args {
		if value, ok := arg.(redact); ok {
			stringArgs = append(stringArgs, value.value)
			shownArgs = append(shownArgs, "<REDACTED>")
		} else {
			stringArgs = append(stringArgs, fmt.Sprintf("%s", arg))
			shownArgs = append(shownArgs, fmt.Sprintf("%s", arg))
		}
	}

	command := exec.Command(r.command, stringArgs...)
	if len(env) > 0 {
		command.Env = append(os.Environ(), env...)
	}

	_, err := r.err.Write([]byte(fmt.Sprintf("\nExecuting: \"%s %s\"\nThis could take a few moments...\n", r.command, strings.Join(shownArgs, " "))))
	if err != nil {
		return nil, nil, err
	}

	command.Stdout = io.MultiWriter(outWriter, &outBufWriter)
	command.Stderr = io.MultiWriter(errWriter, &errBufWriter)

	return &outBufWriter, &errBufWriter, command.Run()
}

type redact struct {
	value string
}

func Redact(a string) redact {
	return redact{a}
}

func (r redact) String() string {
	return r.value
}
