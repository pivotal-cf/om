package api

import (
	"encoding/json"
	"fmt"
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
		return ProductStemcells{}, fmt.Errorf("could not make api request to list stemcells: %s", err)
	}
	defer resp.Body.Close()

	var productStemcells ProductStemcells
	err = json.NewDecoder(resp.Body).Decode(&productStemcells)
	return productStemcells, err
}

func (a Api) AssignStemcell(input ProductStemcells) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = a.sendAPIRequest("PATCH", "/api/v0/stemcell_assignments", jsonData)
	return err
}
