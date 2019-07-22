package generator

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

func NewMetadata(fileBytes []byte) (*Metadata, error) {
	metadata := &Metadata{}
	err := yaml.Unmarshal(fileBytes, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

type Metadata struct {
	Name               string              `yaml:"name"`
	Version            string              `yaml:"product_version"`
	FormTypes          []FormType          `yaml:"form_types"`
	PropertyBlueprints []PropertyBlueprint `yaml:"property_blueprints"`
	JobTypes           []JobType           `yaml:"job_types"`
	PostDeployErrands  []ErrandMetadata    `yaml:"post_deploy_errands"`
	PreDeleteErrands   []ErrandMetadata    `yaml:"pre_delete_errands"`
}

func (m *Metadata) Errands() []ErrandMetadata {
	return append(m.PostDeployErrands, m.PreDeleteErrands...)
}

type ErrandMetadata struct {
	Name string `yaml:"name"`
}

func (m *Metadata) ProductName() string {
	return m.Name
}

func (m *Metadata) ProductVersion() string {
	return m.Version
}

func matchesType(t string) bool {
	switch t {
	case "service_network_az_multi_select", "service_network_az_single_select":
		return true
	}
	return false
}

func (m *Metadata) UsesServiceNetwork() bool {
	for _, job := range m.JobTypes {
		for _, propertyMetadata := range job.PropertyBlueprint {
			if matchesType(propertyMetadata.Type) {
				return true
			}
		}
	}

	for _, propertyMetadata := range m.PropertyBlueprints {
		if matchesType(propertyMetadata.Type) {
			return true
		}
		for _, subPropertyMetadata := range propertyMetadata.PropertyBlueprints {
			if matchesType(subPropertyMetadata.Type) {
				return true
			}
		}
		for _, optionTemplates := range propertyMetadata.OptionTemplates {
			for _, subPropertyMetadata := range optionTemplates.PropertyBlueprints {
				if matchesType(subPropertyMetadata.Type) {
					return true
				}
			}
		}
	}
	return false
}

func (m *Metadata) GetJob(jobName string) (*JobType, error) {
	for _, job := range m.JobTypes {
		if job.Name == jobName {
			return &job, nil
		}
	}
	return nil, fmt.Errorf("job %s not found", jobName)
}

func (m *Metadata) GetPropertyBlueprint(propertyReference string) (*PropertyBlueprint, error) {
	propertyParts := strings.Split(propertyReference, ".")
	jobName := propertyParts[1]
	simplePropertyName := propertyParts[len(propertyParts)-1]

	job, err := m.GetJob(jobName)
	if err == nil {
		return job.GetPropertyBlueprint(propertyReference)
	}

	for _, property := range m.PropertyBlueprints {
		if property.Name == simplePropertyName {
			return &property, nil
		}
	}

	return nil, fmt.Errorf("property %s not found", propertyReference)
}

func (m *Metadata) PropertyInputs() []PropertyInput {
	var propertyInputs []PropertyInput
	for _, form := range m.FormTypes {
		propertyInputs = append(propertyInputs, form.PropertyInputs...)
	}
	return propertyInputs
}
