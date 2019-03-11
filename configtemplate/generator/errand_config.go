package generator

import "fmt"

type Errand struct {
	PostDeployState string `yaml:"post-deploy-state,omitempty"`
	PreDeleteState  string `yaml:"pre-delete-state,omitempty"`
}

func CreateErrandConfig(metadata *Metadata) map[string]Errand {
	errands := make(map[string]Errand)
	for _, errand := range metadata.Errands() {
		errands[errand.Name] = Errand{
			PostDeployState: fmt.Sprintf("((%s_post_deploy_state))", errand.Name),
			PreDeleteState:  fmt.Sprintf("((%s_pre_delete_state))", errand.Name),
		}
	}
	return errands
}

func CreateErrandVars(metadata *Metadata) map[string]interface{} {
	vars := make(map[string]interface{})
	for _, errand := range metadata.Errands() {
		vars[fmt.Sprintf("%s_post_deploy_state", errand.Name)] = "default"
		vars[fmt.Sprintf("%s_pre_delete_state", errand.Name)] = "default"
	}
	return vars
}
