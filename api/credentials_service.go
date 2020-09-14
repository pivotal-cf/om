package api

import (
	"encoding/json"
	"fmt"
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
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/deployed/products/%s/credentials/%s", input.DeployedGUID, input.CredentialReference), nil)
	if err != nil {
		return GetDeployedProductCredentialOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return GetDeployedProductCredentialOutput{}, err
	}

	var credentialOutput GetDeployedProductCredentialOutput
	if err := json.NewDecoder(resp.Body).Decode(&credentialOutput); err != nil {
		return GetDeployedProductCredentialOutput{}, fmt.Errorf("could not unmarshal credentials response: %w", err)
	}

	return credentialOutput, nil
}

func (a Api) ListDeployedProductCredentials(deployedGUID string) (CredentialReferencesOutput, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/deployed/products/%s/credentials", deployedGUID), nil)
	if err != nil {
		return CredentialReferencesOutput{}, fmt.Errorf("could not make api request to credentials endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return CredentialReferencesOutput{}, err
	}

	var credentialReferences CredentialReferencesOutput
	if err := json.NewDecoder(resp.Body).Decode(&credentialReferences); err != nil {
		return CredentialReferencesOutput{}, fmt.Errorf("could not unmarshal credentials response: %w", err)
	}

	return credentialReferences, nil
}
