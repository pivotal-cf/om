package fakes

type Command struct {
	ExecuteCall struct {
		Receives struct {
			Args []string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Command) Execute(args []string) error {
	c.ExecuteCall.Receives.Args = args

	return c.ExecuteCall.Returns.Error
}
