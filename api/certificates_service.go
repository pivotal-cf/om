package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type domainsOutput struct {
	Domains []string `json:"domains"`
}

func (a Api) GenerateCertificate(domains string) (string, error) {
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

	resp, err := a.client.Do(req)
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
