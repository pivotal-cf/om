package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SecurityService struct {
	client httpClient
}

type certResponse struct {
	Cert string `json:"root_ca_certificate_pem"`
}

func NewSecurityService(client httpClient) SecurityService {
	return SecurityService{client: client}
}

func (s SecurityService) FetchRootCACert() (string, error) {
	request, err := http.NewRequest("GET", "/api/v0/security/root_ca_certificate", nil)
	if err != nil {
		return "", fmt.Errorf("failed constructing request: %s", err)
	}

	resp, err := s.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to submit request: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var certResponse certResponse
	err = json.Unmarshal(output, &certResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %s", err)
	}

	return certResponse.Cert, nil
}
