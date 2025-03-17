package api

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

type DeployedProductOutput struct {
	Type string
	GUID string
}

func (a Api) GetDeployedProductManifest(guid string) (string, error) {
	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/deployed/products/%s/manifest", guid), nil)
	if err != nil {
		return "", fmt.Errorf("could not make api request to staged products manifest endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return "", err
	}

	var contents interface{}
	if err := yaml.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return "", fmt.Errorf("could not parse json: %w", err)
	}

	manifest, err := yaml.Marshal(contents)
	if err != nil {
		return "", nil
	}

	return string(manifest), nil
}

func (a Api) ListDeployedProducts() ([]DeployedProductOutput, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/deployed/products", nil)
	if err != nil {
		return []DeployedProductOutput{}, fmt.Errorf("could not make api request to deployed products endpoint: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var deployedProducts []DeployedProductOutput
	if err := json.NewDecoder(resp.Body).Decode(&deployedProducts); err != nil {
		return []DeployedProductOutput{}, fmt.Errorf("could not unmarshal deployed products response: %w", err)
	}

	return deployedProducts, nil
}
