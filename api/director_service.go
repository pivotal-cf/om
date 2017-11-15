package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DirectorService struct {
	client httpClient
}

type DirectorConfiguration struct {
	IAASConfiguration     json.RawMessage `json:"iaas_configuration,omitempty"`
	DirectorConfiguration json.RawMessage `json:"director_configuration,omitempty"`
	SecurityConfiguration json.RawMessage `json:"security_configuration,omitempty"`
}

type NetworkAndAZConfiguration struct {
	NetworkAZ NetworkAndAZFields `json:"network_and_az,omitempty"`
}

type NetworkAndAZFields struct {
	Network     map[string]string `json:"network,omitempty"`
	SingletonAZ map[string]string `json:"singleton_availability_zone,omitempty"`
}

func NewDirectorService(client httpClient) DirectorService {
	return DirectorService{
		client: client,
	}
}

func (d DirectorService) NetworkAndAZ(input NetworkAndAZConfiguration) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not make json: %s", err)
	}

	req, err := http.NewRequest("PUT", "/api/v0/staged/director/network_and_az", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("could not create api request to network and AZ endpoint: %s", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to network and AZ endpoint: %s", err)
	}

	return ValidateStatusOK(resp)
}

func (d DirectorService) Properties(input DirectorConfiguration) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not make json: %s", err)
	}

	req, err := http.NewRequest("PUT", "/api/v0/staged/director/properties", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("could not assigns director configuration properties: %s", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make api request to director properties: %s", err)
	}

	return ValidateStatusOK(resp)
}
