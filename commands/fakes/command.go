package fakes

type Command struct {
	ExecuteCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (c *Command) Execute() error {
	c.ExecuteCall.CallCount++

	return c.ExecuteCall.Returns.Error
}
