package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type CertificateAuthoritiesService struct {
	client httpClient
}

type ActivateCertificateAuthorityInput struct {
	GUID string
}

type DeleteCertificateAuthorityInput struct {
	GUID string
}

type CertificateAuthorityInput struct {
	CertPem       string `json:"cert_pem"`
	PrivateKeyPem string `json:"private_key_pem"`
}

type CertificateAuthoritiesOutput struct {
	CAs []CA `json:"certificate_authorities"`
}

type CA struct {
	GUID      string `json:"guid"`
	Issuer    string `json:"issuer"`
	CreatedOn string `json:"created_on"`
	ExpiresOn string `json:"expires_on"`
	Active    bool   `json:"active"`
	CertPEM   string `json:"cert_pem"`
}

func NewCertificateAuthoritiesService(client httpClient) CertificateAuthoritiesService {
	return CertificateAuthoritiesService{
		client: client,
	}
}

func (c CertificateAuthoritiesService) List() (CertificateAuthoritiesOutput, error) {
	var output CertificateAuthoritiesOutput

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

func (c CertificateAuthoritiesService) Regenerate() error {
	req, err := http.NewRequest("POST", "/api/v0/certificate_authorities/active/regenerate", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
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

func (c CertificateAuthoritiesService) Create(certBody CertificateAuthorityInput) (CA, error) {
	var output CA

	body, err := json.Marshal(certBody)
	if err != nil {
		return CA{}, err // not tested
	}

	req, err := http.NewRequest("POST", "/api/v0/certificate_authorities", bytes.NewReader(body))
	if err != nil {
		return CA{}, err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return CA{}, err
	}

	if err = ValidateStatusOK(resp); err != nil {
		return CA{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return CA{}, err
	}

	return output, nil
}

func (c CertificateAuthoritiesService) Activate(input ActivateCertificateAuthorityInput) error {

	path := fmt.Sprintf("/api/v0/certificate_authorities/%s/activate", input.GUID)

	req, err := http.NewRequest("POST", path, bytes.NewReader([]byte{'{', '}'}))
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (c CertificateAuthoritiesService) Delete(input DeleteCertificateAuthorityInput) error {

	path := fmt.Sprintf("/api/v0/certificate_authorities/%s", input.GUID)

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if err = ValidateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
