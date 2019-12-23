package executor

import (
	"os/exec"
	"strings"
)

type Executor struct {
	omPath string
}

func NewExecutor(omPath string) Executor {
	return Executor{
		omPath: omPath,
	}
}

func (ex Executor) GetDescription(commandName string) (string, error) {
	output, err := ex.RunOmCommand(commandName, "--help")
	if err != nil {
		return "", err
	}

	return strings.Split(string(output), "\n")[1], nil
}

func (ex Executor) GetCommandNames() ([]string, error) {
	output, err := ex.RunOmCommand("--help")
	if err != nil {
		return nil, err
	}

	outputLines := strings.Split(string(output), "\n")

	var isCommand bool
	var commands []string
	for _, commandLine := range outputLines {
		if strings.Contains(commandLine, "Commands:") && !isCommand {
			isCommand = true
			continue
		}

		if isCommand && commandLine != "" {
			splitCommandLine := strings.Fields(commandLine)
			commands = append(commands, splitCommandLine[0])
		}
	}

	return commands, nil
}

func (ex Executor) GetCommandHelp(commandName string) ([]byte, error) {
	return ex.RunOmCommand(commandName, "--help")
}

func (ex Executor) RunOmCommand(args ...string) ([]byte, error) {
	command := exec.Command(ex.omPath, args...)
	command.Dir = "."
	return command.Output()
}
