package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		return fmt.Errorf("could not marshal json: %s", err)
	}

	verb := "PUT"
	endpoint := fmt.Sprintf("/api/v0/staged/vm_extensions/%s", input.Name)
	req, err := http.NewRequest(verb, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("could not create api request %s %s: %s", verb, endpoint, err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send api request to %s %s: %s", verb, endpoint, err.Error())
	}

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var vmExtensions VMExtensionResponse
	if err = json.Unmarshal(body, &vmExtensions); err != nil {
		return nil, fmt.Errorf("could not parse json: %s", err)
	}

	return vmExtensions.VMExtensions, nil
}

func (a Api) DeleteVMExtension(name string) error {
	endpoint := fmt.Sprintf("/api/v0/staged/vm_extensions/%s", name)
	_, err := a.sendAPIRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	return nil
}
