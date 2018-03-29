package proofing

type SelectorPropertyOptionTemplate struct {
	Name               string                    `yaml:"name"`
	SelectValue        string                    `yaml:"select_value"`
	PropertyBlueprints []SimplePropertyBlueprint `yaml:"property_blueprints"`
	NamedManifests     []NamedManifest           `yaml:"named_manifests"`

	// TODO: find_object: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/selector_property_option_template.rb#L11
}
