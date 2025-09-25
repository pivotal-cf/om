package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
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

// StemcellManifest represents the structure of stemcell.MF
// Only relevant fields are included
// Example fields: os, version, infrastructure
type StemcellManifest struct {
	OperatingSystem string `yaml:"operating_system"`
	Version         string `yaml:"version"`
	CloudProperties struct {
		Infrastructure string `yaml:"infrastructure"`
	} `yaml:"cloud_properties"`
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

func extractStemcellManifest(tgzPath string) (StemcellManifest, error) {
	f, err := os.Open(tgzPath)
	if err != nil {
		return StemcellManifest{}, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return StemcellManifest{}, err
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return StemcellManifest{}, err
		}
		if hdr.Name == "stemcell.MF" {
			var mf StemcellManifest
			manifestBytes, err := io.ReadAll(tarReader)
			if err != nil {
				return StemcellManifest{}, err
			}
			err = yaml.Unmarshal(manifestBytes, &mf)
			if err != nil {
				return StemcellManifest{}, err
			}
			return mf, nil
		}
	}
	return StemcellManifest{}, io.EOF // not found
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

	manifest, err := extractStemcellManifest(stemcellFilename)
	if err != nil {
		return false, fmt.Errorf("could not extract stemcell manifest: %w", err)
	}

	osField := manifest.OperatingSystem
	versionField := manifest.Version
	iaasField := manifest.CloudProperties.Infrastructure

	if osField == "" || versionField == "" || iaasField == "" {
		return false, fmt.Errorf("stemcell manifest missing required fields: operating_system, version, or cloud_properties.infrastructure")
	}

	if validVersion {
		for _, stemcell := range report.AvailableStemcells {
			if stemcell.OS == osField && stemcell.Version == versionField && iaasField == report.InfrastructureType {
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
