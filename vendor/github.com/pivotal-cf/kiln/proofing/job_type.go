package proofing

type JobType struct {
	Name          string `yaml:"name"`
	ResourceLabel string `yaml:"resource_label"`
	Description   string `yaml:"description,omitempty"`

	Manifest    string      `yaml:"manifest"`
	MaxInFlight interface{} `yaml:"max_in_flight"`

	Canaries     int  `yaml:"canaries"`
	Serial       bool `yaml:"serial,omitempty"`
	SingleAZOnly bool `yaml:"single_az_only"`

	Errand                     bool `yaml:"errand"`
	RunPreDeleteErrandDefault  bool `yaml:"run_pre_delete_errand_default"`
	RunPostDeployErrandDefault bool `yaml:"run_post_deploy_errand_default"`

	Templates               []Template           `yaml:"templates"`
	InstanceDefinition      InstanceDefinition   `yaml:"instance_definition"`
	ResourceDefinitions     []ResourceDefinition `yaml:"resource_definitions"`
	PropertyBlueprints      PropertyBlueprints   `yaml:"property_blueprints,omitempty"`
	RequiresProductVersions []ProductVersion     `yaml:"requires_product_versions"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/job_type.rb#L11-L15
	// TODO: more validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/job_type.rb#L33-L55
	// TODO: find_object: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/job_type.rb#L57-L58
	// TODO: max_in_flight can be int or percentage
}
