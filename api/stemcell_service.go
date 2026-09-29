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
	"sort"
	"strings"

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
	Name            string      `yaml:"name"`
	OperatingSystem string      `yaml:"operating_system"`
	Version         string      `yaml:"version"`
	Variant         interface{} `yaml:"variant"`
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

// stemcellVariantNamePatterns mirrors how Ops Manager derives a stemcell's
// variant from its name when the stemcell manifest has no explicit variant.
var stemcellVariantNamePatterns = map[string]*regexp.Regexp{
	"fips": regexp.MustCompile(`-fips-`),
}

func variantsFromName(name string) []string {
	variants := []string{}
	for variant, pattern := range stemcellVariantNamePatterns {
		if pattern.MatchString(name) {
			variants = append(variants, variant)
		}
	}
	return variants
}

// variants returns the stemcell variants declared in the manifest, falling back
// to deriving them from the stemcell name (as Ops Manager does).
func (mf stemcellManifest) variants() []string {
	switch v := mf.Variant.(type) {
	case string:
		return nonEmptyStrings([]interface{}{v})
	case []interface{}:
		return nonEmptyStrings(v)
	}
	return variantsFromName(mf.Name)
}

func nonEmptyStrings(values []interface{}) []string {
	result := []string{}
	for _, value := range values {
		if s := strings.TrimSpace(fmt.Sprint(value)); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func sameVariants(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a = append([]string{}, a...)
	b = append([]string{}, b...)
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type stemcellLibraryEntry struct {
	Name    string `json:"name"`
	OS      string `json:"os"`
	Version string `json:"version"`
	// Variant is only reported by Ops Manager 3.3.6+; nil means it was not reported.
	Variant *[]string `json:"variant"`
}

// matchesVariants compares against the declared variants when Ops Manager reports
// them. Otherwise only name-derived variants can be compared: variants declared
// only in a stemcell manifest (e.g. esm) are invisible to such Ops Managers.
func (s stemcellLibraryEntry) matchesVariants(declared, fromName []string) bool {
	if s.Variant != nil {
		return sameVariants(*s.Variant, declared)
	}
	return sameVariants(variantsFromName(s.Name), fromName)
}

func (a Api) listStemcellLibrary() ([]stemcellLibraryEntry, error) {
	resp, err := a.sendAPIRequest("GET", "/api/v0/stemcell_associations", nil)
	if err != nil {
		return nil, fmt.Errorf("could not make api request to list stemcells: %w", err)
	}
	defer resp.Body.Close()

	if err = validateStatusOK(resp); err != nil {
		return nil, err
	}

	var associations struct {
		StemcellLibrary []stemcellLibraryEntry `json:"stemcell_library"`
	}
	err = json.NewDecoder(resp.Body).Decode(&associations)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %s", err)
	}

	return associations.StemcellLibrary, nil
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

func availableStemcellMatches(report DiagnosticReport, os, version, infrastructure string) bool {
	for _, stemcell := range report.AvailableStemcells {
		if stemcell.OS == os && stemcell.Version == version && infrastructureMatches(infrastructure, report.InfrastructureType) {
			return true
		}
	}
	return false
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
		// The diagnostic report does not distinguish stemcell variants (e.g. FIPS),
		// so OS/version matches are confirmed against the stemcell library.
		var library []stemcellLibraryEntry
		libraryLoaded := false
		variantUploaded := func(os, version string, declaredVariants, nameVariants []string) (bool, error) {
			if !libraryLoaded {
				var libraryErr error
				library, libraryErr = a.listStemcellLibrary()
				if libraryErr != nil {
					return false, fmt.Errorf("could not determine stemcell variants on Ops Manager: %w", libraryErr)
				}
				libraryLoaded = true
			}
			for _, stemcell := range library {
				if stemcell.OS == os && stemcell.Version == version && stemcell.matchesVariants(declaredVariants, nameVariants) {
					return true, nil
				}
			}
			return false, nil
		}

		// Try to match by OS, version, and infrastructure from stemcell manifest
		// so that duplicate uploads are avoided regardless of local filename.
		manifest, extractErr := extractStemcellManifest(stemcellFilename)
		if extractErr == nil {
			osField := manifest.OperatingSystem
			versionField := manifest.Version
			iaasField := manifest.CloudProperties.Infrastructure
			if osField != "" && versionField != "" && iaasField != "" &&
				availableStemcellMatches(report, osField, versionField, iaasField) {
				// The manifest is authoritative for the variant, so don't let the
				// filename fallbacks below override it.
				return variantUploaded(osField, versionField, manifest.variants(), variantsFromName(manifest.Name))
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
			parsedVariants := variantsFromName(baseFilename)
			for _, variant := range parsedVariants {
				parsedOS = strings.TrimSuffix(parsedOS, "-"+variant)
			}
			if availableStemcellMatches(report, parsedOS, parsedVersion, parsedInfra) {
				found, err := variantUploaded(parsedOS, parsedVersion, parsedVariants, parsedVariants)
				if err != nil || found {
					return found, err
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
