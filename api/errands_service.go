package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (a Api) UpdateStagedProductErrands(productID string, errandName string, postDeployState interface{}, preDeleteState interface{}) error {
	errandsListOutput := ErrandsListOutput{
		Errands: []Errand{
			{
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

	path := fmt.Sprintf("/api/v0/staged/products/%s/errands", productID)
	_, err = a.sendAPIRequest("PUT", path, payload)
	if err != nil {
		return fmt.Errorf("failed to set errand state: %s", err)
	}

	return nil
}

func (a Api) ListStagedProductErrands(productID string) (ErrandsListOutput, error) {
	var errandsListOutput ErrandsListOutput

	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/staged/products/%s/errands", productID), nil)
	if err != nil {
		return errandsListOutput, fmt.Errorf("failed to list errands: %s", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&errandsListOutput)
	if err != nil {
		return errandsListOutput, err
	}

	return errandsListOutput, nil
}
