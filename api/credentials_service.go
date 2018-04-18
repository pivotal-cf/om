package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GetDeployedProductCredentialInput struct {
	DeployedGUID        string
	CredentialReference string
}

type GetDeployedProductCredentialOutput struct {
	Credential Credential `json:"credential"`
}

type CredentialReferencesOutput struct {
	Credentials []string `json:"credentials"`
}

type Credential struct {
	Type  string            `json:"type"`
	Value map[string]string `json:"value"`
}

func (a Api) GetDeployedProductCredential(input GetDeployedProductCredentialInput) (GetDeployedProductCredentialOutput, error) {
	path := fmt.Sprintf("/api/v0/deployed/products/%s/credentials/%s", input.DeployedGUID, input.CredentialReference)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return GetDeployedProductCredentialOutput{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return GetDeployedProductCredentialOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return GetDeployedProductCredentialOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GetDeployedProductCredentialOutput{}, err
	}

	var credentialOutput GetDeployedProductCredentialOutput
	err = json.Unmarshal(respBody, &credentialOutput)
	if err != nil {
		return GetDeployedProductCredentialOutput{}, fmt.Errorf("could not unmarshal credentials response: %s", err)
	}

	return credentialOutput, nil
}

func (a Api) ListDeployedProductCredentials(deployedGUID string) (CredentialReferencesOutput, error) {
	path := fmt.Sprintf("/api/v0/deployed/products/%s/credentials", deployedGUID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return CredentialReferencesOutput{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return CredentialReferencesOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
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
