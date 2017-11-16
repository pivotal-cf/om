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

type AZConfiguration struct {
	AvailabilityZones json.RawMessage `json:"availability_zones,omitempty"`
}

type NetworkAndAZConfiguration struct {
	NetworkAZ json.RawMessage `json:"network_and_az,omitempty"`
}

type DirectorProperties struct {
	IAASConfiguration     json.RawMessage `json:"iaas_configuration,omitempty"`
	DirectorConfiguration json.RawMessage `json:"director_configuration,omitempty"`
	SecurityConfiguration json.RawMessage `json:"security_configuration,omitempty"`
	SyslogConfiguration   json.RawMessage `json:"syslog_configuration,omitempty"`
}

func NewDirectorService(client httpClient) DirectorService {
	return DirectorService{
		client: client,
	}
}

func (d DirectorService) AZConfiguration(input AZConfiguration) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	return d.sendAPIRequest("PUT", "/api/v0/staged/director/availability_zones", jsonData)
}

func (d DirectorService) NetworksConfiguration(input json.RawMessage) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	return d.sendAPIRequest("PUT", "/api/v0/staged/director/networks", jsonData)
}

func (d DirectorService) NetworkAndAZ(input NetworkAndAZConfiguration) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	return d.sendAPIRequest("PUT", "/api/v0/staged/director/network_and_az", jsonData)
}

func (d DirectorService) Properties(input DirectorProperties) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	return d.sendAPIRequest("PUT", "/api/v0/staged/director/properties", jsonData)
}

func (d DirectorService) sendAPIRequest(verb, endpoint string, jsonData []byte) error {
	req, err := http.NewRequest(verb, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("could not create api request %s %s: %s", verb, endpoint, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send api request to %s %s: %s", verb, endpoint, err.Error())
	}

	return ValidateStatusOK(resp)
}
