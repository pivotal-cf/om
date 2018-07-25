package config

type ProductConfiguration struct {
	ProductProperties        map[string]interface{} `yaml:"product-properties,omitempty"`
	NetworkProperties        map[string]interface{} `yaml:"network-properties,omitempty"`
	ResourceConfigProperties map[string]interface{} `yaml:"resource-config,omitempty"`
}

type VMExtenstionConfig struct {
	VMExtension struct {
		Name            string                 `yaml:"name"`
		CloudProperties map[string]interface{} `yaml:"cloud_properties,omitempty"`
	} `yaml:"vm-extension-config,omitempty"`
}
