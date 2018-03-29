package proofing

type SelectorPropertyInput struct {
	SimplePropertyInput `yaml:",inline"`

	SelectorPropertyInputs []SelectorOptionPropertyInput `yaml:"selector_property_inputs,omitempty"`
}
