package config

type ProductConfiguration struct {
	ProductProperties        map[string]interface{} `yaml:"product-properties,omitempty"`
	NetworkProperties        map[string]interface{} `yaml:"network-properties,omitempty"`
	ResourceConfigProperties map[string]interface{} `yaml:"resource-config,omitempty"`
}
