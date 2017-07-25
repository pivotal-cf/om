package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CredentialOutput struct {
	Credential Credential `json:"credential"`
}

type Credential struct {
	Type  string            `json:"type"`
	Value map[string]string `json:"value"`
}

type CredentialsService struct {
	client   httpClient
	progress progress
}

func NewCredentialsService(client httpClient, progress progress) CredentialsService {
	return CredentialsService{
		client:   client,
		progress: progress,
	}
}

func (cr CredentialsService) Fetch(deployedGUID, credential string) (CredentialOutput, error) {
	path := fmt.Sprintf("/api/v0/deployed/products/%s/credentials/%s", deployedGUID, credential)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return CredentialOutput{}, err
	}

	resp, err := cr.client.Do(req)
	if err != nil {
		return CredentialOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return CredentialOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CredentialOutput{}, err
	}

	var credentialOutput CredentialOutput
	err = json.Unmarshal(respBody, &credentialOutput)
	if err != nil {
		return CredentialOutput{}, fmt.Errorf("could not unmarshal credentials response: %s", err)
	}

	return credentialOutput, nil
}
