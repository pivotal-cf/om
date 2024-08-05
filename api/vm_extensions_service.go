package api

import (
	"encoding/json"
	"fmt"
	"io"
)

type VMExtensionResponse struct {
	VMExtensions []VMExtension `json:"vm_extensions"`
}

type VMExtension struct {
	Name            string                 `yaml:"name" json:"name"`
	CloudProperties map[string]interface{} `yaml:"cloud_properties" json:"cloud_properties"`
}

type CreateVMExtension struct {
	Name            string          `json:"name"`
	CloudProperties json.RawMessage `json:"cloud_properties"`
}

func (a Api) CreateStagedVMExtension(input CreateVMExtension) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}

	resp, err := a.sendAPIRequest("PUT", fmt.Sprintf("/api/v0/staged/vm_extensions/%s", input.Name), jsonData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) ListStagedVMExtensions() ([]VMExtension, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/staged/vm_extensions", nil)
	if err != nil {
		return nil, err // un-tested
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var vmExtensions VMExtensionResponse
	if err = json.Unmarshal(body, &vmExtensions); err != nil {
		return nil, fmt.Errorf("could not parse json: %w", err)
	}

	return vmExtensions.VMExtensions, nil
}

func (a Api) DeleteVMExtension(name string) error {
	resp, err := a.sendAPIRequest("DELETE", fmt.Sprintf("/api/v0/staged/vm_extensions/%s", name), nil)
	if validateError := validateStatusOK(resp); validateError != nil {
		return validateError
	}

	return err
}
