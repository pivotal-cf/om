package fakes

import "github.com/pivotal-cf/om/api"

type SetupService struct {
	SetupCall struct {
		CallCount int
		Receives  struct {
			Input api.SetupInput
		}
		Returns struct {
			Output api.SetupOutput
			Error  error
		}
	}

	EnsureAvailabilityCall struct {
		CallCount int
		Receives  struct {
			Input api.EnsureAvailabilityInput
		}
		Returns struct {
			Outputs []api.EnsureAvailabilityOutput
			Errors  []error
		}
	}
}

func (ss *SetupService) Setup(input api.SetupInput) (api.SetupOutput, error) {
	ss.SetupCall.Receives.Input = input
	ss.SetupCall.CallCount++

	return ss.SetupCall.Returns.Output, ss.SetupCall.Returns.Error
}

func (ss *SetupService) EnsureAvailability(input api.EnsureAvailabilityInput) (api.EnsureAvailabilityOutput, error) {
	ss.EnsureAvailabilityCall.Receives.Input = input

	if len(ss.EnsureAvailabilityCall.Returns.Outputs) <= ss.EnsureAvailabilityCall.CallCount {
		ss.EnsureAvailabilityCall.Returns.Outputs = append(ss.EnsureAvailabilityCall.Returns.Outputs, api.EnsureAvailabilityOutput{})
	}

	if len(ss.EnsureAvailabilityCall.Returns.Errors) <= ss.EnsureAvailabilityCall.CallCount {
		ss.EnsureAvailabilityCall.Returns.Errors = append(ss.EnsureAvailabilityCall.Returns.Errors, nil)
	}

	output := ss.EnsureAvailabilityCall.Returns.Outputs[ss.EnsureAvailabilityCall.CallCount]
	err := ss.EnsureAvailabilityCall.Returns.Errors[ss.EnsureAvailabilityCall.CallCount]

	ss.EnsureAvailabilityCall.CallCount++

	return output, err
}
