package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type CertificatesService struct {
	client httpClient
}

type domainsOutput struct {
	Domains []string `json:"domains"`
}

func NewCertificatesService(client httpClient) CertificatesService {
	return CertificatesService{
		client: client,
	}
}

func (c CertificatesService) Generate(domains string) (string, error) {
	domainsOutput := domainsOutput{
		Domains: strings.Split(domains, ","),
	}

	payload, err := json.Marshal(domainsOutput)
	if err != nil {
		return "", err // not tested
	}

	req, err := http.NewRequest("POST", "/api/v0/certificates/generate", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}
