package fakes

import "github.com/pivotal-cf/om/api"

type SetupService struct {
	SetupCall struct {
		Receives struct {
			Input api.SetupInput
		}
		Returns struct {
			Output api.SetupOutput
			Error  error
		}
	}
}

func (ss *SetupService) Setup(input api.SetupInput) (api.SetupOutput, error) {
	ss.SetupCall.Receives.Input = input

	return ss.SetupCall.Returns.Output, ss.SetupCall.Returns.Error
}
