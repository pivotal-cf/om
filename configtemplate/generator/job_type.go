package generator

import (
	"fmt"
	"strings"
)

type JobType struct {
	Name                string               `yaml:"name"`
	PropertyMetadata    []PropertyMetadata   `yaml:"property_blueprints"`
	ResourceDefinitions []ResourceDefinition `yaml:"resource_definitions"`
	InstanceDefinition  InstanceDefinition   `yaml:"instance_definition"`
}

func (j *JobType) InstanceDefinitionConfigurable() bool {
	return j.InstanceDefinition.Configurable
}

func (j *JobType) IsIncluded() bool {
	if j.InstanceDefinition.Default == 0 && !j.InstanceDefinition.Configurable {
		return false
	}
	return true
}

type InstanceDefinition struct {
	Configurable bool `yaml:"configurable"`
	Default      int  `yaml:"default"`
}

type ResourceDefinition struct {
	Configurable bool        `yaml:"configurable"`
	Default      interface{} `yaml:"default"`
	Name         string      `yaml:"name"`
	Type         string      `yaml:"type"`
}

func (j *JobType) HasPersistentDisk() bool {
	for _, resourceDef := range j.ResourceDefinitions {
		if resourceDef.Name == "persistent_disk" && resourceDef.Configurable {
			return true
		}
	}
	return false
}

func (j *JobType) GetPropertyMetadata(propertyName string) (*PropertyMetadata, error) {
	propertyParts := strings.Split(propertyName, ".")
	simplePropertyName := propertyParts[len(propertyParts)-1]

	for _, property := range j.PropertyMetadata {
		if property.Name == simplePropertyName {
			return &property, nil
		}
	}
	return nil, fmt.Errorf("property %s not found on job %s", propertyName, j.Name)
}
