package proofing

type SelectorOptionPropertyInput struct {
	Reference string `yaml:"reference"`
	Label     string `yaml:"label"`

	PropertyInputs []SimplePropertyInput `yaml:"property_inputs,omitempty"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/selector_option_property_input.rb#L8-L12
}
