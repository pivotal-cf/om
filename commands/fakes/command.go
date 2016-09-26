package fakes

import "github.com/pivotal-cf/om/commands"

type Command struct {
	ExecuteCall struct {
		Receives struct {
			Args []string
		}
		Returns struct {
			Error error
		}
	}

	UsageCall struct {
		Returns struct {
			Usage commands.Usage
		}
	}
}

func (c *Command) Execute(args []string) error {
	c.ExecuteCall.Receives.Args = args

	return c.ExecuteCall.Returns.Error
}

func (c *Command) Usage() commands.Usage {
	return c.UsageCall.Returns.Usage
}
