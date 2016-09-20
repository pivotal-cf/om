package fakes

type Helper struct {
	HelpCall struct {
		Returns struct {
			Help string
		}
	}
}

func (h *Helper) Help() string {
	return h.HelpCall.Returns.Help
}
