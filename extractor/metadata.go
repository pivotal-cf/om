package extractor

type Metadata struct {
	Name             string `yaml:"name"`
	Version          string `yaml:"product_version"`
	StemcellCriteria struct {
		OS                   string `yaml:"os"`
		Version              string `yaml:"version"`
		PatchSecurityUpdates bool   `yaml:"enable_patch_security_updates"`
	} `yaml:"stemcell_criteria"`
	Raw []byte
}
