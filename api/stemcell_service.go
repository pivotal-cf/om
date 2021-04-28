package api

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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
		return ProductStemcells{}, fmt.Errorf("could not make api request to list stemcells: %w", err)
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
		return fmt.Errorf("could not marshal json: %w", err)
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

func (a Api) CheckStemcellAvailability(stemcellFilename string) (bool, error) {
	report, err := a.GetDiagnosticReport()
	if err != nil {
		return false, fmt.Errorf("failed to get diagnostic report: %s", err)
	}

	info, err := a.Info()
	if err != nil {
		return false, fmt.Errorf("cannot retrieve version of Ops Manager: %w", err)
	}

	validVersion, err := info.VersionAtLeast(2, 6)
	if err != nil {
		return false, fmt.Errorf("could not determine version was 2.6+ compatible: %s", err)
	}

	if validVersion {
		for _, stemcell := range report.AvailableStemcells {
			if stemcell.Filename == filepath.Base(stemcellFilename) {
				return true, nil
			}
		}
	}

	for _, stemcell := range report.Stemcells {
		if stemcell == filepath.Base(stemcellFilename) {
			return true, nil
		}
	}

	return false, nil
}
