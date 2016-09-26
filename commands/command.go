package commands

type Command interface {
	Execute(args []string) error
	Usage() Usage
}
