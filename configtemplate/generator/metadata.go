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
	Name              string             `yaml:"name"`
	Version           string             `yaml:"product_version"`
	FormTypes         []FormType         `yaml:"form_types"`
	PropertyMetadata  []PropertyMetadata `yaml:"property_blueprints"`
	JobTypes          []JobType          `yaml:"job_types"`
	ProvidesVersions  []ProvidesVersion  `yaml:"provides_product_versions"`
	PostDeployErrands []ErrandMetadata   `yaml:"post_deploy_errands"`
	PreDeleteErrands  []ErrandMetadata   `yaml:"pre_delete_errands"`
}

func (m *Metadata) Errands() []ErrandMetadata {
	return append(m.PostDeployErrands, m.PreDeleteErrands...)
}

type ErrandMetadata struct {
	Name string `yaml:"name"`
}

type ProvidesVersion struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (m *Metadata) ProductName() string {
	return m.Name
}

func (m *Metadata) ProductVersion() string {
	if len(m.ProvidesVersions) > 0 {
		return m.ProvidesVersions[0].Version
	} else {
		return m.Version
	}
}

func (m *Metadata) UsesServiceNetwork() bool {
	for _, job := range m.JobTypes {
		for _, propertyMetadata := range job.PropertyMetadata {
			if "service_network_az_multi_select" == propertyMetadata.Type || "service_network_az_single_select" == propertyMetadata.Type {
				return true
			}
		}
	}

	for _, propertyMetadata := range m.PropertyMetadata {
		if "service_network_az_multi_select" == propertyMetadata.Type || "service_network_az_single_select" == propertyMetadata.Type {
			return true
		}
		for _, subPropertyMetadata := range propertyMetadata.PropertyMetadata {
			if "service_network_az_multi_select" == subPropertyMetadata.Type || "service_network_az_single_select" == subPropertyMetadata.Type {
				return true
			}
		}
		for _, optionTemplates := range propertyMetadata.OptionTemplates {
			for _, subPropertyMetadata := range optionTemplates.PropertyMetadata {
				if "service_network_az_multi_select" == subPropertyMetadata.Type || "service_network_az_single_select" == subPropertyMetadata.Type {
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
	return nil, fmt.Errorf("Job %s not found", jobName)
}

func (m *Metadata) GetPropertyMetadata(propertyName string) (*PropertyMetadata, error) {
	propertyParts := strings.Split(propertyName, ".")
	jobName := propertyParts[1]
	simplePropertyName := propertyParts[len(propertyParts)-1]

	job, err := m.GetJob(jobName)
	if err == nil {
		return job.GetPropertyMetadata(propertyName)
	}

	for _, property := range m.PropertyMetadata {
		if property.Name == simplePropertyName {
			return &property, nil
		}
	}

	return nil, fmt.Errorf("Property %s not found", propertyName)
}

func (m *Metadata) Properties() []Property {
	var properties []Property
	for _, form := range m.FormTypes {
		properties = append(properties, form.Properties...)
	}
	return properties
}
