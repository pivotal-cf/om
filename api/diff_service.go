package api

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type ProductDiff struct {
	Manifest       ManifestDiff         `json:"manifest"`
	RuntimeConfigs []RuntimeConfigsDiff `json:"runtime_configs"`
}

type ManifestDiff struct {
	Status string `json:"status"`
	Diff   string `json:"diff"`
}

type RuntimeConfigsDiff struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Diff   string `json:"diff"`
}

func (a *Api) ProductDiff(productName string) (string, error) {
	productGUID, err := a.checkStagedProducts(productName)
	if err != nil {
		return "", err
	}

	if productGUID == "" {
		return "", fmt.Errorf(`could not find product "%s"`, productName)
	}

	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/products/%s/diff", productGUID), nil)
	if err != nil {
		return "", errors.Wrap(err, "could not request product diff")
	}

	err = validateStatusOK(resp)
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve product diff")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read response body for product diff")
	}

	var diff ProductDiff
	if err = json.Unmarshal(body, &diff); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not unmarshal product diff response: %s", string(body)))
	}

	return diff.Manifest.Diff, nil
}
