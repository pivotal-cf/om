package proofing

type Variable struct {
	Name    string      `yaml:"name"`
	Options interface{} `yaml:"options,omitempty"` // TODO: schema?
	Type    string      `yaml:"type"`
}
