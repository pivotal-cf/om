package api

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type ProductStemcells struct {
	Products []ProductStemcell `json:"products"`
}

type ProductStemcell struct {
	GUID                    string   `json:"guid,omitempty"`
	ProductName             string   `json:"identifier,omitempty"`
	StagedForDeletion       bool     `json:"is_staged_for_deletion,omitempty"`
	StagedStemcellVersion   string   `json:"staged_stemcell_version,omitempty"`
	RequiredStemcellVersion string   `json:"required_stemcell_version,omitempty"`
	AvailableVersions       []string `json:"available_stemcell_versions,omitempty"`
}

func (a Api) ListStemcells() (ProductStemcells, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/stemcell_assignments", nil)
	if err != nil {
		return ProductStemcells{}, errors.Wrap(err, "could not make api request to list stemcells")
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return ProductStemcells{}, err
	}

	var productStemcells ProductStemcells
	err = json.NewDecoder(resp.Body).Decode(&productStemcells)
	if err != nil {
		return ProductStemcells{}, fmt.Errorf("invalid JSON: %s", err)
	}

	return productStemcells, nil
}

func (a Api) AssignStemcell(input ProductStemcells) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return errors.Wrap(err, "could not marshal json")
	}

	resp, err := a.sendAPIRequest("PATCH", "/api/v0/stemcell_assignments", jsonData)
	if err != nil {
		return err
	}

	if err = validateStatusOK(resp); err != nil {
		return err
	}

	return nil
}
