package commands

type Command interface {
	Execute() error
}
