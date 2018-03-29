package jhanda

import "fmt"

// CommandSet is a structural collection of executable commands referenced by a
// name. Use this object to translate the name of the command given at the
// commandline to an executable object to be invoked.
type CommandSet map[string]Command

// Execute will invoke the Execute method of the Command that matches the name
// provided as "command", passing "args". Execute will return an error in the
// case that the command cannot be found by the given name.
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

// Usage will return the Usage object of the Command that matches the name
// provided as "command". Usage will return an error in the case that the
// command cannot be found by the given name.
func (cs CommandSet) Usage(command string) (Usage, error) {
	cmd, ok := cs[command]
	if !ok {
		return Usage{}, fmt.Errorf("unknown command: %s", command)
	}

	return cmd.Usage(), nil
}
