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

func (ex Executor) GetCommandNamesAndDescriptions() (map[string]string, error) {
	output, err := ex.RunOmCommand("--help")
	if err != nil {
		return nil, err
	}

	outputLines := strings.Split(string(output), "\n")

	var inTheCommandZone bool
	commands := make(map[string]string)
	for _, commandLine := range outputLines {
		if strings.Contains(commandLine, "Commands:") && !inTheCommandZone {
			inTheCommandZone = true
			continue
		}

		if strings.Contains(commandLine, "Global Flags:") {
			inTheCommandZone = false
			break
		}

		if inTheCommandZone && commandLine != "" {
			splitCommandLine := strings.Fields(commandLine)
			commands[splitCommandLine[0]] = strings.Join(splitCommandLine[1:], " ")
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
