package api

import (
	"encoding/json"
	"fmt"
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

	resp, err := a.sendAPIRequest("GET", "/api/v0/certificate_authorities", nil)
	if err != nil {
		return output, err
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return CertificateAuthoritiesOutput{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return CertificateAuthoritiesOutput{}, err
	}

	return output, nil
}

func (a Api) RegenerateCertificates() error {
	resp, err := a.sendAPIRequest("POST", "/api/v0/certificate_authorities/active/regenerate", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) GenerateCertificateAuthority() (CA, error) {
	var output CA

	resp, err := a.sendAPIRequest("POST", "/api/v0/certificate_authorities/generate", nil)
	if err != nil {
		return CA{}, err
	}

	defer resp.Body.Close()

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

	resp, err := a.sendAPIRequest("POST", "/api/v0/certificate_authorities", body)
	if err != nil {
		return CA{}, err
	}

	defer resp.Body.Close()

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
	resp, err := a.sendAPIRequest("POST", fmt.Sprintf("/api/v0/certificate_authorities/%s/activate", input.GUID), []byte("{}"))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}

func (a Api) DeleteCertificateAuthority(input DeleteCertificateAuthorityInput) error {
	path := fmt.Sprintf("/api/v0/certificate_authorities/%s", input.GUID)
	resp, err := a.sendAPIRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
