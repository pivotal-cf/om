package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CredentialReferencesOutput struct {
	Credentials []string `json:"credentials"`
}

type CredentialReferencesService struct {
	client   httpClient
	progress progress
}

func NewCredentialReferencesService(client httpClient, progress progress) CredentialReferencesService {
	return CredentialReferencesService{
		client:   client,
		progress: progress,
	}
}

func (cr CredentialReferencesService) List(deployedGUID string) (CredentialReferencesOutput, error) {
	path := fmt.Sprintf("/api/v0/deployed/products/%s/credentials", deployedGUID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return CredentialReferencesOutput{}, err
	}

	resp, err := cr.client.Do(req)
	if err != nil {
		return CredentialReferencesOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return CredentialReferencesOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CredentialReferencesOutput{}, err
	}

	var credentialReferences CredentialReferencesOutput
	err = json.Unmarshal(respBody, &credentialReferences)
	if err != nil {
		return CredentialReferencesOutput{}, fmt.Errorf("could not unmarshal credentials response: %s", err)
	}

	return credentialReferences, nil
}
