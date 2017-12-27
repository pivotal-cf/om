package jhanda

import "fmt"

type CommandSet map[string]Command

func (cs CommandSet) Execute(command string, args []string) error {
	cmd, ok := cs[command]
	if !ok {
		return fmt.Errorf("unknown command: %s", command)
	}

	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "-help" {
			return cs.Execute("help", []string{command})
		}
	}

	err := cmd.Execute(args)
	if err != nil {
		return fmt.Errorf("could not execute %q: %s", command, err)
	}

	return nil
}

func (cs CommandSet) Usage(command string) (Usage, error) {
	cmd, ok := cs[command]
	if !ok {
		return Usage{}, fmt.Errorf("unknown command: %s", command)
	}

	return cmd.Usage(), nil
}
