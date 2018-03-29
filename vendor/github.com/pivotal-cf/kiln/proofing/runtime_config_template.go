package proofing

type RuntimeConfigTemplate struct {
	Name          string `yaml:"name"`
	RuntimeConfig string `yaml:"runtime_config"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/runtime_config_template.rb#L7-L12
}
