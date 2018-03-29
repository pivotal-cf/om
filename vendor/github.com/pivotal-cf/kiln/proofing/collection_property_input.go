package proofing

type CollectionPropertyInput struct {
	SimplePropertyInput `yaml:",inline"`

	PropertyInputs []CollectionSubfieldPropertyInput `yaml:"property_inputs"`
}
