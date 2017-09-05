package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var readAll = ioutil.ReadAll

type Errand struct {
	Name       string      `json:"name"`
	PostDeploy interface{} `json:"post_deploy,omitempty"`
	PreDelete  interface{} `json:"pre_delete,omitempty"`
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

func (es ErrandsService) SetState(productID string, errandName string, postDeployState interface{}, preDeleteState interface{}) error {
	path := fmt.Sprintf("/api/v0/staged/products/%s/errands", productID)

	errandsListOutput := ErrandsListOutput{
		Errands: []Errand{
			Errand{
				Name:       errandName,
				PostDeploy: postDeployState,
				PreDelete:  preDeleteState,
			},
		},
	}

	payload, err := json.Marshal(errandsListOutput)
	if err != nil {
		return err // not tested
	}

	req, err := http.NewRequest("PUT", path, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := es.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		rawBody, err := readAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("failed to set errand state: %d %s", resp.StatusCode, string(rawBody))
	}

	return nil
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
