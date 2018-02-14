package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type VMExtensionsService struct {
	client httpClient
}

type CreateVMExtension struct {
	Name            string          `json:"name"`
	CloudProperties json.RawMessage `json:"cloud_properties"`
}

func NewVMExtensionsService(client httpClient) VMExtensionsService {
	return VMExtensionsService{
		client: client,
	}
}

type VMExtensionInput struct {
	Name            string `json:"name"`
	CloudProperties string `json:"cloud_properties"`
}

func (v VMExtensionsService) Create(input CreateVMExtension) error {
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

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send api request to %s %s: %s", verb, endpoint, err.Error())
	}

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
