package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
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

// stemcellManifest represents the structure of stemcell.MF inside a stemcell .tgz.
// Only relevant fields for duplicate detection are included.
type stemcellManifest struct {
	OperatingSystem string `yaml:"operating_system"`
	Version         string `yaml:"version"`
	CloudProperties struct {
		Infrastructure string `yaml:"infrastructure"`
	} `yaml:"cloud_properties"`
}

func extractStemcellManifest(tgzPath string) (stemcellManifest, error) {
	f, err := os.Open(tgzPath)
	if err != nil {
		return stemcellManifest{}, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return stemcellManifest{}, err
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stemcellManifest{}, err
		}
		if hdr.Name == "stemcell.MF" {
			var mf stemcellManifest
			manifestBytes, err := io.ReadAll(tarReader)
			if err != nil {
				return stemcellManifest{}, err
			}
			err = yaml.Unmarshal(manifestBytes, &mf)
			if err != nil {
				return stemcellManifest{}, err
			}
			return mf, nil
		}
	}
	return stemcellManifest{}, fmt.Errorf("stemcell.MF not found in %s", tgzPath)
}

// parseStemcellFilename attempts to extract OS, version, and infrastructure from a stemcell filename.
// Handles both common formats:
// - bosh-stemcell-1.1250-vsphere-esxi-ubuntu-jammy-go_agent.tgz
// - light-bosh-stemcell-2022.4-google-kvm-windows2022-go_agent.tgz
// - bosh-vsphere-esxi-ubuntu-jammy-go_agent-1.1250.tgz
func parseStemcellFilename(filename string) (os string, version string, infrastructure string) {
	// Extract just the filename if a path is provided
	filename = filepath.Base(filename)
	filename = strings.TrimSuffix(filename, ".tgz")

	// Try pattern 1: [light-]bosh-stemcell-VERSION-INFRASTRUCTURE-OS-go_agent
	// The light- prefix is optional (used for some stemcells like Windows)
	re1 := regexp.MustCompile(`^(?:light-)?bosh-stemcell-(\d+\.\d+(?:\.\d+)?)-(.+?)-((?:ubuntu|centos|rhel|windows|alpine|debian).+?)-go_agent$`)
	if matches := re1.FindStringSubmatch(filename); matches != nil {
		version = matches[1]
		infrastructure = matches[2]
		os = matches[3]
		return
	}

	// Try pattern 2: bosh-INFRASTRUCTURE-OS-go_agent-VERSION
	re2 := regexp.MustCompile(`^bosh-(.+?)-((?:ubuntu|centos|rhel|windows|alpine|debian).+?)-go_agent-(\d+\.\d+(?:\.\d+)?)$`)
	if matches := re2.FindStringSubmatch(filename); matches != nil {
		infrastructure = matches[1]
		os = matches[2]
		version = matches[3]
		return
	}

	return "", "", ""
}

// infrastructureMatches reports whether the stemcell's infrastructure matches the
// report's infrastructure type. Treats equivalent infrastructure types as matches:
// - "warden" and "docker" are equivalent (for Docker Ops Manager)
// - "vsphere" variants (vsphere, vsphere-esxi) are equivalent
func infrastructureMatches(manifestInfrastructure, reportInfrastructureType string) bool {
	if manifestInfrastructure == reportInfrastructureType {
		return true
	}
	// Docker equivalence
	if (manifestInfrastructure == "warden" && reportInfrastructureType == "docker") ||
		(manifestInfrastructure == "docker" && reportInfrastructureType == "warden") {
		return true
	}
	// vSphere variants are equivalent (vsphere, vsphere-esxi)
	manifestIsVsphere := manifestInfrastructure == "vsphere" ||
		manifestInfrastructure == "vsphere-esxi"
	reportIsVsphere := reportInfrastructureType == "vsphere" ||
		reportInfrastructureType == "vsphere-esxi"
	return manifestIsVsphere && reportIsVsphere
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
		// Try to match by OS, version, and infrastructure from stemcell manifest
		// so that duplicate uploads are avoided regardless of local filename.
		manifest, extractErr := extractStemcellManifest(stemcellFilename)
		if extractErr == nil {
			osField := manifest.OperatingSystem
			versionField := manifest.Version
			iaasField := manifest.CloudProperties.Infrastructure
			if osField != "" && versionField != "" && iaasField != "" {
				for _, stemcell := range report.AvailableStemcells {
					if stemcell.OS == osField && stemcell.Version == versionField && infrastructureMatches(iaasField, report.InfrastructureType) {
						return true, nil
					}
				}
			}
		}
		// Fall back to exact filename match when manifest cannot be used (e.g. file not found, invalid tgz)
		baseFilename := filepath.Base(stemcellFilename)
		for _, stemcell := range report.AvailableStemcells {
			if stemcell.Filename == baseFilename {
				return true, nil
			}
		}

		// Try smart filename parsing when exact filename doesn't match.
		// This handles cases where the requested filename and available filename have different formats.
		parsedOS, parsedVersion, parsedInfra := parseStemcellFilename(baseFilename)
		if parsedOS != "" && parsedVersion != "" && parsedInfra != "" {
			for _, stemcell := range report.AvailableStemcells {
				if stemcell.OS == parsedOS && stemcell.Version == parsedVersion && infrastructureMatches(parsedInfra, report.InfrastructureType) {
					return true, nil
				}
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
