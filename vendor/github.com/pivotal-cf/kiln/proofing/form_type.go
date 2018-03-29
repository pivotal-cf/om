package proofing

type FormType struct {
	Verifiers      []VerifierBlueprint `yaml:"verifiers,omitempty"`
	PropertyInputs PropertyInputs      `yaml:"property_inputs"`

	Name        string `yaml:"name"`
	Label       string `yaml:"label"`
	Description string `yaml:"description"`
	Markdown    string `yaml:"markdown,omitempty"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/form_type.rb#L13-L24
}
