package commands

import "fmt"

type Set map[string]Command

func (s Set) Execute(command string, args []string) error {
	cmd, ok := s[command]
	if !ok {
		return fmt.Errorf("unknown command: %s", command)
	}

	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "-help" {
			return s.Execute("help", []string{command})
		}
	}

	err := cmd.Execute(args)
	if err != nil {
		return fmt.Errorf("could not execute %q: %s", command, err)
	}

	return nil
}

func (s Set) Usage(command string) (Usage, error) {
	cmd, ok := s[command]
	if !ok {
		return Usage{}, fmt.Errorf("unknown command: %s", command)
	}

	return cmd.Usage(), nil
}
