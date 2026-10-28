package validator

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateInstallationZip recognizes an Ops Manager installation export by
// its installation.yml, metadata.json, root postgres dump, or metadata/ and
// deployments/ files. This is format recognition only, not an integrity check.
func ValidateInstallationZip(path string) error {
	zipper, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("file: \"%s\" is not a valid zip file", path)
	}
	defer zipper.Close()

	for _, f := range zipper.File {
		if f.Name == "installation.yml" ||
			f.Name == "metadata.json" ||
			(!strings.Contains(f.Name, "/") && filepath.Ext(f.Name) == ".postgres" && !f.FileInfo().IsDir()) ||
			((strings.HasPrefix(f.Name, "metadata/") || strings.HasPrefix(f.Name, "deployments/")) && !strings.HasSuffix(f.Name, "/") && !f.FileInfo().IsDir()) {
			return nil
		}
	}

	return fmt.Errorf("file: \"%s\" is not a valid installation file (expected root installation.yml, root metadata.json, a root *.postgres file, or a file under metadata/ or deployments/). Validate that the provided installation file is correct, or run \"om export-installation\" and try again.", path)
}
