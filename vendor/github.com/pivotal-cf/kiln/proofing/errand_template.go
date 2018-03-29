package proofing

type ErrandTemplate struct {
	Name        string   `yaml:"name"`
	Colocated   bool     `yaml:"colocated"`
	RunDefault  bool     `yaml:"run_default"`
	Instances   []string `yaml:"instances"` // TODO: how to validate?
	Label       string   `yaml:"label"`
	Description string   `yaml:"description"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/errand_template.rb#L11-L22
}
