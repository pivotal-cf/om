package proofing

type CollectionPropertyBlueprint struct {
	SimplePropertyBlueprint `yaml:",inline"`

	PropertyBlueprints []SimplePropertyBlueprint `yaml:"property_blueprints"`
	NamedManifests     []NamedManifest           `yaml:"named_manifests"`
}
