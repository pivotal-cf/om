package api

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type ProductMultiStemcells struct {
	Products []ProductMultiStemcell `json:"products"`
}

type StemcellObject struct {
	OS      string `json:"os,omitempty"`
	Version string `json:"version,omitempty"`
}

type ProductMultiStemcell struct {
	GUID              string           `json:"guid,omitempty"`
	ProductName       string           `json:"identifier,omitempty"`
	StagedForDeletion bool             `json:"is_staged_for_deletion,omitempty"`
	StagedStemcells   []StemcellObject `json:"staged_stemcells,omitempty"`
	RequiredStemcells []StemcellObject `json:"required_stemcells,omitempty"`
	AvailableVersions []StemcellObject `json:"available_stemcells,omitempty"`
}

func (a Api) ListMultiStemcells() (ProductMultiStemcells, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/stemcell_associations", nil)
	if err != nil {
		return ProductMultiStemcells{}, errors.Wrap(err, "could not make api request to list stemcells")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return ProductMultiStemcells{}, err
	}

	var productStemcells ProductMultiStemcells
	err = json.NewDecoder(resp.Body).Decode(&productStemcells)
	if err != nil {
		return ProductMultiStemcells{}, fmt.Errorf("invalid JSON: %s", err)
	}

	return productStemcells, nil
}

func (a Api) AssignMultiStemcell(input ProductMultiStemcells) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return errors.Wrap(err, "could not marshal json")
	}

	resp, err := a.sendAPIRequest("PATCH", "/api/v0/stemcell_associations", jsonData)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
