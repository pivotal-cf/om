package validator_test

import (
	"archive/zip"
	"io/fs"
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

	DescribeTable("accepts installation export data", func(name string, mode fs.FileMode) {
		path := createInstallationZipEntries(tempDir, []installationZipEntry{{name: name, mode: mode}})

		Expect(validator.ValidateInstallationZip(path)).To(Succeed())
	},
		Entry("a root postgres dump", "rails_database_dump.postgres", fs.FileMode(0)),
		Entry("a file under metadata", "metadata/p.yml", fs.FileMode(0)),
		Entry("a nested file under metadata", "metadata/sub/p.yml", fs.FileMode(0)),
		Entry("a file under deployments", "deployments/bosh-state.json", fs.FileMode(0)),
		Entry("a zero-byte file under metadata", "metadata/empty", fs.FileMode(0)),
	)

	DescribeTable("rejects entries that are not installation markers", func(name string, mode fs.FileMode) {
		path := createInstallationZipEntries(tempDir, []installationZipEntry{{name: name, mode: mode}})

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidInstallationMessage(path)))
	},
		Entry("an uppercase postgres suffix", "DB.POSTGRES", fs.FileMode(0)),
		Entry("a dot slash postgres path", "./db.postgres", fs.FileMode(0)),
		Entry("a dot slash metadata path", "./metadata/p.yml", fs.FileMode(0)),
		Entry("an uppercase metadata path", "METADATA/p.yml", fs.FileMode(0)),
		Entry("a nested postgres file", "x/db.postgres", fs.FileMode(0)),
		Entry("a postgres directory", "db.postgres/", fs.ModeDir|0755),
		Entry("a metadata directory", "metadata/", fs.ModeDir|0755),
		Entry("a mode-marked metadata directory without a slash", "metadata/sub", fs.ModeDir|0755),
		Entry("a mode-marked deployments directory without a slash", "deployments/sub", fs.ModeDir|0755),
	)

	It("rejects a zip containing only metadata and deployments directory entries", func() {
		path := createInstallationZipEntries(tempDir, []installationZipEntry{
			{name: "metadata/", mode: fs.ModeDir | 0755},
			{name: "deployments/", mode: fs.ModeDir | 0755},
		})

		Expect(validator.ValidateInstallationZip(path)).To(MatchError(invalidInstallationMessage(path)))
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
	zipEntries := make([]installationZipEntry, 0, len(entries))
	for name, contents := range entries {
		zipEntries = append(zipEntries, installationZipEntry{name: name, contents: contents})
	}
	return createInstallationZipEntries(dir, zipEntries)
}

type installationZipEntry struct {
	name     string
	contents string
	mode     fs.FileMode
}

func createInstallationZipEntries(dir string, entries []installationZipEntry) string {
	path := filepath.Join(dir, "installation.zip")
	file, err := os.Create(path)
	Expect(err).To(Succeed())

	writer := zip.NewWriter(file)
	for _, zipEntry := range entries {
		header := &zip.FileHeader{Name: zipEntry.name}
		if zipEntry.mode != 0 {
			header.SetMode(zipEntry.mode)
		}
		entry, err := writer.CreateHeader(header)
		Expect(err).To(Succeed())
		_, err = entry.Write([]byte(zipEntry.contents))
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
	return `file: "` + path + `" is not a valid installation file (expected root installation.yml, root metadata.json, a root *.postgres file, or a file under metadata/ or deployments/). Validate that the provided installation file is correct, or run "om export-installation" and try again.`
}
