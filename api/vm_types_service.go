package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

type CreateVMTypes struct {
	VMTypes []CreateVMType `json:"vm_types" yaml:"vm_types"`
}

type VMTypesResponse struct {
	VMTypes []VMType `json:"vm_types"`
}

type CreateVMType struct {
	RAM             uint                   `yaml:"ram"`
	Name            string                 `yaml:"name"`
	CPU             uint                   `yaml:"cpu"`
	EphemeralDisk   uint                   `yaml:"ephemeral_disk"`
	ExtraProperties map[string]interface{} `yaml:",inline"`
}

type VMType struct {
	CreateVMType
	BuiltIn bool `json:"builtin" yaml:"builtin"`
}

func (c *CreateVMType) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	c.ExtraProperties = make(map[string]interface{})

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	for key, value := range raw {
		var u float64
		var ok bool
		switch key {
		case "name":
			c.Name = fmt.Sprintf("%v", value)

		case "ram":
			if u, ok = value.(float64); !ok {
				return fmt.Errorf("could not marshal ram into uint")
			}

			c.RAM = uint(u)

		case "cpu":
			if u, ok = value.(float64); !ok {
				return fmt.Errorf("could not marshal cpu into uint")
			}

			c.CPU = uint(u)

		case "ephemeral_disk":
			if u, ok = value.(float64); !ok {
				return fmt.Errorf("could not marshal ephemeral_disk into uint")
			}

			c.EphemeralDisk = uint(u)
		default:
			if key != "builtin" {
				c.ExtraProperties[key] = value
			}
		}
	}

	return nil
}

func (v *VMType) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}

	c := CreateVMType{}
	c.UnmarshalJSON(b)

	v.CreateVMType = c

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	if builtin, ok := raw["builtin"]; ok {
		if _, assertOk := builtin.(bool); assertOk {
			v.BuiltIn = builtin.(bool)
		}
	}

	return nil
}

func (c CreateVMType) MarshalJSON() ([]byte, error) {
	raw := make(map[string]interface{})
	raw["name"] = c.Name
	raw["ram"] = c.RAM
	raw["cpu"] = c.CPU
	raw["ephemeral_disk"] = c.EphemeralDisk
	for k, v := range c.ExtraProperties {
		raw[k] = v
	}

	return json.Marshal(raw)
}

func (a Api) CreateCustomVMTypes(input CreateVMTypes) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return errors.Wrap(err, "could not marshal json")
	}

	resp, err := a.sendAPIRequest("PUT", "/api/v0/vm_types", jsonData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) ListVMTypes() ([]VMType, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/vm_types", nil)
	if err != nil {
		return nil, err // un-tested
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var vmTypes VMTypesResponse
	if err = json.Unmarshal(body, &vmTypes); err != nil {
		return nil, errors.Wrap(err, "could not parse json")
	}

	return vmTypes.VMTypes, nil
}

func (a Api) DeleteCustomVMTypes() error {
	resp, err := a.sendAPIRequest("DELETE", "/api/v0/vm_types", nil)
	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return err
}
