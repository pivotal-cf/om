package vmlifecyclecommands

import (
	"fmt"
	"io"
)

//go:generate counterfeiter -o ./fakes/taskModifierService.go --fake-name TaskModifierService . taskModifierService
type taskModifierService interface {
	ModifyTasksWithSecrets(stderr io.Writer, taskDir string, configPaths []string, varsFiles []string) error
}

type PrepareTasksWithSecrets struct {
	stdout    io.Writer
	stderr    io.Writer
	service   taskModifierService
	TaskDir   string   `long:"task-dir"   description:"Directory containing the task files to be modified in place" required:"true"`
	ConfigDir []string `long:"config-dir" description:"Directory containing the config files with secrets wrapped in double parentheses" required:"true"`
	VarDir    []string `long:"var-dir" description:"Directory containing the variable files with secrets to ignore"`
}

func NewSecretsModifierCommand(stdout, stderr io.Writer, service taskModifierService) PrepareTasksWithSecrets {
	return PrepareTasksWithSecrets{
		stdout:  stdout,
		stderr:  stderr,
		service: service,
	}
}
func (c *PrepareTasksWithSecrets) Execute(args []string) error {
	err := c.service.ModifyTasksWithSecrets(c.stderr, c.TaskDir, c.ConfigDir, c.VarDir)
	if err != nil {
		return fmt.Errorf("could not modify tasks: %s", err)
	}

	_, err = fmt.Fprintln(c.stdout, "successfully added secrets to provided tasks")
	if err != nil {
		return err
	}

	return nil
}
