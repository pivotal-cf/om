package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

type DeployedProductOutput struct {
	Type string
	GUID string
}

type DeployedProductsService struct {
	client httpClient
}

func NewDeployedProductsService(client httpClient) DeployedProductsService {
	return DeployedProductsService{
		client: client,
	}
}

func (s DeployedProductsService) Manifest(guid string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v0/deployed/products/%s/manifest", guid), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not make api request to staged products manifest endpoint: %s", err)
	}

	if err = ValidateStatusOK(resp); err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var contents interface{}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = yaml.Unmarshal(body, &contents); err != nil {
		return "", fmt.Errorf("could not parse json: %s", err)
	}

	manifest, err := yaml.Marshal(contents)
	if err != nil {
		return "", err // this should never happen, all valid json can be marshalled
	}

	return string(manifest), nil
}

func (s DeployedProductsService) List() ([]DeployedProductOutput, error) {
	req, err := http.NewRequest("GET", "/api/v0/deployed/products", nil)
	if err != nil {
		return []DeployedProductOutput{}, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return []DeployedProductOutput{}, fmt.Errorf("could not make api request to deployed products endpoint: %s", err)
	}
	defer resp.Body.Close()

	if err = ValidateStatusOK(resp); err != nil {
		return []DeployedProductOutput{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []DeployedProductOutput{}, err
	}

	var deployedProducts []DeployedProductOutput
	err = json.Unmarshal(respBody, &deployedProducts)
	if err != nil {
		return []DeployedProductOutput{}, fmt.Errorf("could not unmarshal deployed products response: %s", err)
	}

	return deployedProducts, nil
}
