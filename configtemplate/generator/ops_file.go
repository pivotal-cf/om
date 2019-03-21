package generator

import "fmt"

type Ops struct {
	Type  string       `yaml:"type"`
	Path  string       `yaml:"path"`
	Value OpsValueType `yaml:"value,omitempty"`
}

type OpsNameValue struct {
	Value string `yaml:"name"`
}

func (n *OpsNameValue) Parameters() []string {
	return []string{n.Value}
}

type OpsValue struct {
	Value          string `yaml:"value"`
	SelectedOption string `yaml:"selected_option"`
}

func (n *OpsValue) Parameters() []string {
	return []string{n.Value}
}

type OpsValueType interface {
	Parameters() []string
}

type StringOpsValue string

func (n StringOpsValue) Parameters() []string {
	return []string{fmt.Sprintf("%v", n)}
}
