package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
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

	response, err := s.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to submit request: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		out, err := httputil.DumpResponse(response, true)
		if err != nil {
			return "", fmt.Errorf("request failed: unexpected response: %s", err)
		}
		return "", fmt.Errorf("could not make api request: unexpected response.\n%s", out)
	}

	output, err := ioutil.ReadAll(response.Body)
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
