package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func (s DeployedProductsService) DeployedProducts() ([]DeployedProductOutput, error) {
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
