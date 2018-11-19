package config

type ProductConfiguration struct {
	ProductName              string                  `yaml:"product-name,omitempty"`
	ProductProperties        map[string]interface{}  `yaml:"product-properties,omitempty"`
	NetworkProperties        map[string]interface{}  `yaml:"network-properties,omitempty"`
	ResourceConfigProperties map[string]interface{}  `yaml:"resource-config,omitempty"`
	ErrandConfigs            map[string]ErrandConfig `yaml:"errand-config,omitempty"`
	Field                    map[string]interface{}  `yaml:",inline"`
}

type ErrandConfig struct {
	PostDeployState interface{} `yaml:"post-deploy-state,omitempty"`
	PreDeleteState  interface{} `yaml:"pre-delete-state,omitempty"`
}

type VMExtensionConfig struct {
	VMExtension struct {
		Name            string                 `yaml:"name"`
		CloudProperties map[string]interface{} `yaml:"cloud_properties,omitempty"`
	} `yaml:"vm-extension-config,omitempty"`
}
