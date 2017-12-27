package jhanda

//go:generate counterfeiter -o ./internal/fakes/command.go --fake-name Command . Command
type Command interface {
	Execute(args []string) error
	Usage() Usage
}
