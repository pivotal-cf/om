package vmmanagers

import (
	"fmt"
	"gopkg.in/go-playground/validator.v9"
)

type Status int

const (
	Success Status = iota
	Exist
	StateMismatch
	Unknown
	Incomplete
)

type StateInfo struct {
	IAAS string `yaml:"iaas"`
	ID   string `yaml:"vm_id"`
}

type OpsmanConfig struct {
	Vsphere   *VsphereConfig         `yaml:"vsphere,omitempty"`
	GCP       *GCPConfig             `yaml:"gcp,omitempty"`
	AWS       *AWSConfig             `yaml:"aws,omitempty"`
	Azure     *AzureConfig           `yaml:"azure,omitempty"`
	Openstack *OpenstackConfig       `yaml:"openstack,omitempty"`
	Unknown   map[string]interface{} `yaml:",inline"`
}

type OpsmanConfigFilePayload struct {
	OpsmanConfig OpsmanConfig           `yaml:"opsman-configuration"`
	Fields       map[string]interface{} `yaml:",inline"`
}

func validateIAASConfig(config interface{}) error {
	validate := validator.New()
	err := validate.Struct(config)

	return checkFormatedError("validation of Config failed: %s", err)
}

func checkFormatedError(format string, err error) error {
	if err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}
