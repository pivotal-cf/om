package config

type ProductConfiguration struct {
	ProductProperties        map[string]interface{}  `yaml:"product-properties,omitempty"`
	NetworkProperties        map[string]interface{}  `yaml:"network-properties,omitempty"`
	ResourceConfigProperties map[string]interface{}  `yaml:"resource-config,omitempty"`
	ErrandConfigs            map[string]ErrandConfig `yaml:"errand-config,omitempty"`
}

type ErrandConfig struct {
	PostDeployState interface{} `yaml:"post-deploy-state,omitempty"`
	PreDeleteState  interface{} `yaml:"pre-delete-state,omitempty"`
}
