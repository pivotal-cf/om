package api

import (
	"encoding/json"
	"io/ioutil"
)

type DomainsInput struct {
	Domains []string `json:"domains"`
}

func (a Api) GenerateCertificate(domains DomainsInput) (string, error) {
	payload, err := json.Marshal(domains)
	if err != nil {
		return "", err // not tested
	}

	resp, err := a.sendAPIRequest("POST", "/api/v0/certificates/generate", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
