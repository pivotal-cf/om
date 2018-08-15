package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type SSLCertificateInput struct {
	CertPem       string `json:"certificate"`
	PrivateKeyPem string `json:"private_key"`
}

type SSLCertificateOutput struct {
	Certificate SSLCertificate `json:"ssl_certificate"`
}

type SSLCertificate struct {
	Certificate string `json:"certificate"`
}

func (a Api) UpdateSSLCertificate(certBody SSLCertificateInput) error {
	body, err := json.Marshal(certBody)
	if err != nil {
		return err // not tested
	}

	req, err := http.NewRequest("PUT", "/api/v0/settings/ssl_certificate", bytes.NewReader(body))
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

func (a Api) GetSSLCertificate() (SSLCertificateOutput, error) {
	var output SSLCertificateOutput

	req, err := http.NewRequest("GET", "/api/v0/settings/ssl_certificate", nil)
	if err != nil {
		return output, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return output, err
	}

	if err = validateStatusOK(resp); err != nil {
		return SSLCertificateOutput{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&output)
	if err != nil {
		return output, err
	}

	if output.Certificate.Certificate == "" {
		output.Certificate.Certificate = "Ops Manager Self Signed Cert"
	}

	return output, nil
}

func (a Api) DeleteSSLCertificate() error {
	req, err := http.NewRequest("DELETE", "/api/v0/settings/ssl_certificate", nil)
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
