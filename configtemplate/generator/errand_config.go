package generator

import "fmt"

type Errand struct {
	PostDeployState string `yaml:"post-deploy-state,omitempty"`
	PreDeleteState  string `yaml:"pre-delete-state,omitempty"`
}

func CreateErrandConfig(metadata *Metadata) map[string]Errand {
	errands := make(map[string]Errand)
	for _, errand := range metadata.PostDeployErrands {
		errands[errand.Name] = Errand{
			PostDeployState: fmt.Sprintf("((%s_post_deploy_state))", errand.Name),
			PreDeleteState:  "",
		}
	}

	for _, errand := range metadata.PreDeleteErrands {
		errands[errand.Name] = Errand{
			PostDeployState: "",
			PreDeleteState:  fmt.Sprintf("((%s_pre_delete_state))", errand.Name),
		}
	}

	return errands
}

func CreateErrandVars(metadata *Metadata) map[string]interface{} {
	vars := make(map[string]interface{})
	for _, errand := range metadata.PostDeployErrands {
		vars[fmt.Sprintf("%s_post_deploy_state", errand.Name)] = "default"
	}

	for _, errand := range metadata.PreDeleteErrands {
		vars[fmt.Sprintf("%s_pre_delete_state", errand.Name)] = "default"
	}

	return vars
}
