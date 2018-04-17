package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (a Api) ListCertificateAuthorities() (CertificateAuthoritiesOutput, error) {
	var output CertificateAuthoritiesOutput

	req, err := http.NewRequest("GET", "/api/v0/certificate_authorities", nil)
	if err != nil {
		return output, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return output, err
	}

	if err = validateStatusOK(resp); err != nil {
		return CertificateAuthoritiesOutput{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, err
	}

	return output, nil
}

func (a Api) RegenerateCertificates() error {
	req, err := http.NewRequest("POST", "/api/v0/certificate_authorities/active/regenerate", nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) GenerateCertificateAuthority() (CA, error) {
	var output CA

	req, err := http.NewRequest("POST", "/api/v0/certificate_authorities/generate", nil)
	if err != nil {
		return CA{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return CA{}, err
	}

	if err = validateStatusOK(resp); err != nil {
		return CA{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return CA{}, err
	}

	return output, nil
}

func (a Api) CreateCertificateAuthority(certBody CertificateAuthorityInput) (CA, error) {
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

	resp, err := a.client.Do(req)
	if err != nil {
		return CA{}, err
	}

	if err = validateStatusOK(resp); err != nil {
		return CA{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return CA{}, err
	}

	return output, nil
}

func (a Api) ActivateCertificateAuthority(input ActivateCertificateAuthorityInput) error {

	path := fmt.Sprintf("/api/v0/certificate_authorities/%s/activate", input.GUID)

	req, err := http.NewRequest("POST", path, bytes.NewReader([]byte{'{', '}'}))
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) DeleteCertificateAuthority(input DeleteCertificateAuthorityInput) error {

	path := fmt.Sprintf("/api/v0/certificate_authorities/%s", input.GUID)

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return err // not tested
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
