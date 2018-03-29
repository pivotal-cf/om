package jhanda

//go:generate counterfeiter -o ./internal/fakes/command.go --fake-name Command . Command

// Command defines the interface for a command object type. For an object to be
// executable as a command, you will need to implement these methods.
type Command interface {
	Execute(args []string) error
	Usage() Usage
}
