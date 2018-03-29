package proofing

import "fmt"

type ProductTemplate struct {
	Name                     string `yaml:"name"`
	ProductVersion           string `yaml:"product_version"`
	MinimumVersionForUpgrade string `yaml:"minimum_version_for_upgrade"`
	Label                    string `yaml:"label"`
	Rank                     int    `yaml:"rank"`
	MetadataVersion          string `yaml:"metadata_version"`
	OriginalMetadataVersion  string `yaml:"original_metadata_version"`
	ServiceBroker            bool   `yaml:"service_broker"`

	IconImage           string `yaml:"icon_image"`
	DeprecatedTileImage string `yaml:"deprecated_tile_image"`

	Cloud   string `yaml:"cloud"`
	Network string `yaml:"network"`

	Serial               bool                  `yaml:"serial"`
	InstallTimeVerifiers []InstallTimeVerifier `yaml:"install_time_verifiers"` // TODO: schema?

	BaseReleasesURL         string                  `yaml:"base_releases_url"`
	Variables               []Variable              `yaml:"variables"` // TODO: schema?
	Releases                []Release               `yaml:"releases"`
	StemcellCriteria        StemcellCriteria        `yaml:"stemcell_criteria"`
	PropertyBlueprints      PropertyBlueprints      `yaml:"property_blueprints"`
	FormTypes               []FormType              `yaml:"form_types"`
	JobTypes                []JobType               `yaml:"job_types"`
	RequiresProductVersions []ProductVersion        `yaml:"requires_product_versions"`
	PostDeployErrands       []ErrandTemplate        `yaml:"post_deploy_errands"`
	PreDeleteErrands        []ErrandTemplate        `yaml:"pre_delete_errands"`
	RuntimeConfigs          []RuntimeConfigTemplate `yaml:"runtime_configs"`

	// TODO: validates_presence_of: https://github.com/pivotal-cf/installation/blob/b7be08d7b50d305c08d520ee0afe81ae3a98bd9d/web/app/models/persistence/metadata/product_template.rb#L20-L25
	// TODO: version_attribute: https://github.com/pivotal-cf/installation/blob/b7be08d7b50d305c08d520ee0afe81ae3a98bd9d/web/app/models/persistence/metadata/product_template.rb#L30-L32
	// TODO: validates_string: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L56
	// TODO: validates_integer: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L60
	// TODO: validates_manifest: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L61
	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L64-L70
	// TODO: validates: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L72
	// TODO: validates_object(s): https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L74-L82
	// TODO: find_object: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/product_template.rb#L84-L86
}

func (pt ProductTemplate) AllPropertyBlueprints() []NormalizedPropertyBlueprint {
	var propertyBlueprints []NormalizedPropertyBlueprint

	propertyBlueprints = make([]NormalizedPropertyBlueprint, 0, len(pt.PropertyBlueprints))

	for _, pb := range pt.PropertyBlueprints {
		propertyBlueprints = append(propertyBlueprints, pb.Normalize(".properties")...)
	}

	for _, jobType := range pt.JobTypes {
		for _, pb := range jobType.PropertyBlueprints {
			prefix := fmt.Sprintf(".%s", jobType.Name)
			propertyBlueprints = append(propertyBlueprints, pb.Normalize(prefix)...)
		}
	}

	return propertyBlueprints
}
