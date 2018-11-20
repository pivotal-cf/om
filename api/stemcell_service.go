package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	req, err := http.NewRequest("GET", "/api/v0/stemcell_assignments", nil)
	if err != nil {
		return ProductStemcells{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return ProductStemcells{}, fmt.Errorf("could not make api request to list stemcells: %s", err)
	}

	defer resp.Body.Close()
	err = validateStatusOK(resp)
	if err != nil {
		return ProductStemcells{}, err
	}

	var productStemcells ProductStemcells
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&productStemcells)
	if err != nil {
		return ProductStemcells{}, err
	}
	return productStemcells, nil
}

func (a Api) AssignStemcell(input ProductStemcells) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = a.sendAPIRequest("PATCH", "/api/v0/stemcell_assignments", jsonData)

	return err
}
