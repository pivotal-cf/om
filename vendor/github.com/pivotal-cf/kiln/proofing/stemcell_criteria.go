package proofing

type StemcellCriteria struct {
	OS                         string `yaml:"os"`
	Version                    string `yaml:"version"`
	EnablePatchSecurityUpdates bool   `yaml:"enable_patch_security_updates"`

	// TODO: version_attribute: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/stemcell_criteria.rb#L8-L9
	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/stemcell_criteria.rb#L11-L15
}
