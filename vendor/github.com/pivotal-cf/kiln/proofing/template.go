package proofing

type Template struct {
	Name     string `yaml:"name"`
	Release  string `yaml:"release"`
	Manifest string `yaml:"manifest,omitempty"`
	Consumes string `yaml:"consumes,omitempty"`
	Provides string `yaml:"provides,omitempty"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/template.rb#L9
}
