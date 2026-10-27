package validator

import (
	"archive/zip"
	"fmt"
)

// ValidateInstallationZip checks that path is a zip whose root contains an
// Ops Manager installation marker (installation.yml or metadata.json).
func ValidateInstallationZip(path string) error {
	zipper, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("file: \"%s\" is not a valid zip file", path)
	}
	defer zipper.Close()

	for _, f := range zipper.File {
		if f.Name == "installation.yml" || f.Name == "metadata.json" {
			return nil
		}
	}

	return fmt.Errorf("file: \"%s\" is not a valid installation file (expected installation.yml or metadata.json at the archive root). Validate that the provided installation file is correct, or run \"om export-installation\" and try again.", path)
}
