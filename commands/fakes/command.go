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

	HelpCall struct {
		Returns struct {
			Help string
		}
	}
}

func (c *Command) Execute(args []string) error {
	c.ExecuteCall.Receives.Args = args

	return c.ExecuteCall.Returns.Error
}

func (c *Command) Help() string {
	return c.HelpCall.Returns.Help
}
