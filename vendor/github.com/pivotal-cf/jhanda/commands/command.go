package commands

//go:generate counterfeiter -o ./fakes/command.go --fake-name Command . Command
type Command interface {
	Execute(args []string) error
	Usage() Usage
}
