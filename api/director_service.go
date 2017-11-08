package api

import (
	"fmt"
	"net/http"
	"strings"
)

type DirectorService struct {
	client httpClient
}

func NewDirectorService(client httpClient) DirectorService {
	return DirectorService{
		client: client,
	}
}

func (d DirectorService) NetworkAndAZ(jsonBody string) error {
	req, err := http.NewRequest("PUT", "/api/v0/staged/director/network_and_az", strings.NewReader(jsonBody))
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

func (d DirectorService) Properties(jsonBody string) error {
	req, err := http.NewRequest("PUT", "/api/v0/staged/director/properties", strings.NewReader(jsonBody))
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