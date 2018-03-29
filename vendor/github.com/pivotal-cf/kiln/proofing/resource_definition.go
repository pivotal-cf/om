package proofing

type ResourceDefinition struct {
	Name         string      `yaml:"name"`
	Default      int         `yaml:"default"`
	Configurable bool        `yaml:"configurable"`
	Constraints  interface{} `yaml:"constraints,omitempty"` // TODO: schema?

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/resource_definition.rb#L9
}
