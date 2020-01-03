package api

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type DirectorDiff struct {
	Manifest       ManifestDiff         `json:"manifest"`
	CloudConfig    ManifestDiff         `json:"cloud_config"`
	RuntimeConfigs []RuntimeConfigsDiff `json:"runtime_configs"`
	CPIConfigs     []CPIConfigsDiff     `json:"cpi_configs"`
}

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

type CPIConfigsDiff struct {
	GUID                  string `json:"guid"`
	IAASConfigurationName string `json:"iaas_configuration_name"`
	Status                string `json:"status"`
	Diff                  string `json:"diff"`
}

func (a Api) DirectorDiff() (DirectorDiff, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/director/diff", nil)
	if err != nil {
		return DirectorDiff{}, fmt.Errorf("could not request director diff: %s", err)
	}

	err = validateStatusOK(resp)
	if err != nil {
		return DirectorDiff{}, errors.Wrap(err, "could not retrieve director diff")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DirectorDiff{}, err
	}

	var diff DirectorDiff
	if err = json.Unmarshal(body, &diff); err != nil {
		return DirectorDiff{}, errors.Wrap(err, fmt.Sprintf("could not unmarshal director diff response: %s", string(body)))
	}

	return diff, nil
}

func (a Api) ProductDiff(productName string) (ProductDiff, error) {
	productGUID, err := a.checkStagedProducts(productName)
	if err != nil {
		return ProductDiff{}, err
	}

	if productGUID == "" {
		return ProductDiff{}, fmt.Errorf(`could not find product "%s": it may be invalid, not yet be staged, or be marked for deletion`, productName)
	}

	resp, err := a.sendAPIRequest("GET", fmt.Sprintf("/api/v0/products/%s/diff", productGUID), nil)
	if err != nil {
		return ProductDiff{}, errors.Wrap(err, "could not request product diff")
	}

	err = validateStatusOK(resp)
	if err != nil {
		return ProductDiff{}, errors.Wrap(err, "could not retrieve product diff")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ProductDiff{}, errors.Wrap(err, "could not read response body for product diff")
	}

	var diff ProductDiff
	if err = json.Unmarshal(body, &diff); err != nil {
		return ProductDiff{}, errors.Wrap(err, fmt.Sprintf("could not unmarshal product diff response: %s", string(body)))
	}

	return diff, nil
}
