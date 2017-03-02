package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Errand struct {
	Name       string `json:"name"`
	PostDeploy bool   `json:"post_deploy"`
}

type ErrandsListOutput struct {
	Errands []Errand `json:"errands"`
}

type ErrandsService struct {
	Client httpClient
}

func NewErrandsService(client httpClient) ErrandsService {
	return ErrandsService{Client: client}
}

func (es ErrandsService) List(productID string) (ErrandsListOutput, error) {
	var errandsListOutput ErrandsListOutput

	path := fmt.Sprintf("/api/v0/staged/products/%s/errands", productID)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return errandsListOutput, err
	}

	resp, err := es.Client.Do(req)
	if err != nil {
		return errandsListOutput, err
	}

	err = json.NewDecoder(resp.Body).Decode(&errandsListOutput)
	if err != nil {
		return errandsListOutput, err
	}

	return errandsListOutput, nil
}
