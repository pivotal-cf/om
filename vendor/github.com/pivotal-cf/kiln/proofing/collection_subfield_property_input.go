package proofing

type CollectionSubfieldPropertyInput struct {
	SimplePropertyInput `yaml:",inline"`

	Slug bool `yaml:"slug"`
}
