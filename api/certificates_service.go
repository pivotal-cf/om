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

	respBody, err := ioutil.ReadAll(resp.Body)
	return string(respBody), err
}
