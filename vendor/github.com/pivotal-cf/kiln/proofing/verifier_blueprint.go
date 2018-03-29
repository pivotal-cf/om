package proofing

type VerifierBlueprint struct {
	Name       string      `yaml:"name"`
	Properties interface{} `yaml:"properties"` // TODO: schema?
}
