package proofing

type SimplePropertyInput struct {
	Reference   string `yaml:"reference"`
	Label       string `yaml:"label"`
	Description string `yaml:"description,omitempty"`
	Placeholder string `yaml:"placeholder,omitempty"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/property_input.rb#L9-L13
}
