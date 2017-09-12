package api

import (
	"encoding/json"
	"net/http"
)

type CertificateAuthoritiesService struct {
	client httpClient
}

type CertificateAuthoritiesServiceOutput struct {
	CAs []CA `json:"certificate_authorities"`
}

type CA struct {
	GUID      string
	Issuer    string
	CreatedOn string `json:"created_on"`
	ExpiresOn string `json:"expires_on"`
	Active    bool
	CertPEM   string `json:"cert_pem"`
}

func NewCertificateAuthoritiesService(client httpClient) CertificateAuthoritiesService {
	return CertificateAuthoritiesService{
		client: client,
	}
}

func (c CertificateAuthoritiesService) List() (CertificateAuthoritiesServiceOutput, error) {
	var output CertificateAuthoritiesServiceOutput

	req, err := http.NewRequest("GET", "/api/v0/certificate_authorities", nil)
	if err != nil {
		return output, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return output, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, err
	}

	return output, nil
}

func (c CertificateAuthoritiesService) Generate() (CA, error) {
	var output CA

	req, err := http.NewRequest("POST", "/api/v0/certificate_authorities/generate", nil)
	if err != nil {
		return CA{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return CA{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return CA{}, err
	}

	return output, nil
}
