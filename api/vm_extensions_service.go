package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateVMExtension struct {
	Name            string          `json:"name"`
	CloudProperties json.RawMessage `json:"cloud_properties"`
}

type VMExtensionInput struct {
	Name            string `json:"name"`
	CloudProperties string `json:"cloud_properties"`
}

func (a Api) CreateStagedVMExtension(input CreateVMExtension) error {
	jsonData, err := json.Marshal(&input)

	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	verb := "POST"
	endpoint := "/api/v0/staged/vm_extensions"
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
