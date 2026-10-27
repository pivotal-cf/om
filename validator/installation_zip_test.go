package validator_test

import (
	"archive/zip"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/om/validator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateInstallationZip", func() {
	var tempDir string

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
	})

	It("accepts a zip with installation.yml at its root", func() {
		path := createInstallationZip(tempDir, map[string]string{"installation.yml": "installation"})

		Expect(validator.ValidateInstallationZip(path)).To(Succeed())
	})

	It("accepts a zip with metadata.json at its root", func() {
		path := createInstallationZip(tempDir, map[string]string{"metadata.json": `{"version":"11.0.0-build.451"}`})

		Expect(validator.ValidateInstallationZip(path)).To(Succeed())
	})

	It("rejects a nested metadata.json", func() {
		path := createInstallationZip(tempDir, map[string]string{"foo/metadata.json": `{}`})

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidInstallationMessage(path)))
	})

	It("rejects a nested installation.yml", func() {
		path := createInstallationZip(tempDir, map[string]string{"foo/installation.yml": "installation"})

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidInstallationMessage(path)))
	})

	It("rejects a zip without an installation marker", func() {
		path := createInstallationZip(tempDir, map[string]string{"foo.txt": "content"})

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidInstallationMessage(path)))
	})

	It("rejects a non-zip file", func() {
		path := filepath.Join(tempDir, "not-a-zip")
		Expect(os.WriteFile(path, []byte("plain text"), 0644)).To(Succeed())

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidZipMessage(path)))
	})

	It("rejects a nonexistent path", func() {
		path := filepath.Join(tempDir, "missing.zip")

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidZipMessage(path)))
	})
})

func createInstallationZip(dir string, entries map[string]string) string {
	path := filepath.Join(dir, "installation.zip")
	file, err := os.Create(path)
	Expect(err).To(Succeed())

	writer := zip.NewWriter(file)
	for name, contents := range entries {
		entry, err := writer.Create(name)
		Expect(err).To(Succeed())
		_, err = entry.Write([]byte(contents))
		Expect(err).To(Succeed())
	}
	Expect(writer.Close()).To(Succeed())
	Expect(file.Close()).To(Succeed())

	return path
}

func invalidZipMessage(path string) string {
	return `file: "` + path + `" is not a valid zip file`
}

func invalidInstallationMessage(path string) string {
	return `file: "` + path + `" is not a valid installation file (expected installation.yml or metadata.json at the archive root). Validate that the provided installation file is correct, or run "om export-installation" and try again.`
}
