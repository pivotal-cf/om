package proofing

type InstallTimeVerifier struct {
	Ignorable  bool        `yaml:"ignorable,omitempty"`
	Name       string      `yaml:"name"`
	Properties interface{} `yaml:"properties"` // TODO: schema?
}
