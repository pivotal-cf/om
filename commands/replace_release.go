package commands

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ReplaceRelease struct {
	logger
	Options struct {
		Product       string `long:"product"               required:"true"`
		ProductOutput string `long:"product-output"        required:"true"`
		NewVersion    string `long:"new-product-version"   required:"true"`

		ExistingRelease string `long:"existing-bosh-release" required:"true"`
		NewRelease      string `long:"new-bosh-release"      required:"true"`

		Verbose bool `long:"verbose" default:"false"`
	}
}

func NewReplaceRelease(logger logger) *ReplaceRelease {
	return &ReplaceRelease{
		logger: logger,
	}
}

func (cmd ReplaceRelease) Execute(_ []string) error {
	return replaceRelease(cmd.logger,
		cmd.Options.Verbose,
		cmd.Options.ProductOutput,
		cmd.Options.Product,
		cmd.Options.NewRelease,
		cmd.Options.ExistingRelease,
		cmd.Options.NewVersion,
	)
}

func replaceRelease(logger logger, logDiff bool, outputTileName, inputTilePath, newReleasePath, existingReleaseID, outputTileVersion string) error {
	newReleaseMetadata, err := createReleaseMetadataFromReleaseTarball(newReleasePath)
	if err != nil {
		return fmt.Errorf("failed to generate new release metadata from release tarball: %w", err)
	}
	logger.Printf("calculated new BOSH release metadata: %#v", newReleaseMetadata)

	existingReleaseMetadata, err := queryReleaseMetadataFromTile(inputTilePath, existingReleaseID)
	if err != nil {
		return fmt.Errorf("failed to find expected release metadata in tile for %s: %w", existingReleaseID, err)
	}
	logger.Printf("found existing BOSH release metadata: %#v", existingReleaseMetadata)

	updatedMetadata, err := updateTileMetadata(logger, logDiff, inputTilePath, outputTileVersion, existingReleaseMetadata, newReleaseMetadata)
	if err != nil {
		return fmt.Errorf("failed to generate a new tile manifest: %w", err)
	}
	logger.Println("updated tile metadata")

	err = createTileWithNewMetadataAndReleaseTarball(logger, outputTileName, inputTilePath, newReleasePath, existingReleaseMetadata, updatedMetadata)
	if err != nil {
		return fmt.Errorf("failed to copy or update files in tile: %w", err)
	}
	logger.Println("created new tile")

	return err
}

func updateTileMetadata(logger logger, logDiff bool, inputTilePath, outputTileVersion string, oldRelease, newRelease boshReleaseMetadata) ([]byte, error) {
	metadata, err := readMetadataFromFile(inputTilePath)
	if err != nil {
		return nil, err
	}
	ms := string(metadata)
	var root yaml.Node
	err = yaml.Unmarshal(metadata, &root)
	if err != nil {
		panic(err)
	}
	iterateMap(root.Content[0], func(key, value *yaml.Node) {
		switch key.Value {
		case "releases":
			for _, release := range value.Content {
				nameMatches, versionMatches := false, false
				iterateMap(release, func(key, value *yaml.Node) {
					switch key.Value {
					case "name":
						nameMatches = value.Value == oldRelease.Name
					case "version":
						versionMatches = value.Value == oldRelease.Version
					}
				})
				if nameMatches && versionMatches {
					err = release.Encode(newRelease)
					if err != nil {
						panic(err)
					}
				}
			}
		case "provides_product_versions":
			for _, productVersion := range value.Content {
				iterateMap(productVersion, func(key, value *yaml.Node) {
					if key.Value != "version" {
						return
					}
					value.SetString(outputTileVersion)
				})
			}
		case "product_version":
			value.SetString(outputTileVersion)
		}
	})
	result, err := yaml.Marshal(root.Content[0])
	if err != nil {
		return nil, err
	}
	if logDiff {
		_ = displayDiff(logger, formatYAML(ms), string(result))
	}
	return result, err
}

func formatYAML(in string) string {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(in), &node)
	if err != nil {
		panic(err)
	}
	formatted, err := yaml.Marshal(node.Content[0])
	if err != nil {
		panic(err)
	}
	return string(formatted)
}

func createTileWithNewMetadataAndReleaseTarball(logger logger, outputTilePath, inputTilePath, releaseTarballPath string, oldRelease boshReleaseMetadata, updatedMetadata []byte) error {
	inZip, err := zip.OpenReader(inputTilePath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputTilePath)
	if err != nil {
		return err
	}
	defer closeAndIgnoreError(outFile)
	outZip := zip.NewWriter(outFile)
	defer closeAndIgnoreError(outZip)

	replacedRelease := false
	for _, file := range inZip.File {
		switch file.Name {
		case path.Join("releases", oldRelease.File):
			logger.Printf("\tadding updated release to new tile")
			err = addFileToZip(outZip, path.Join("releases", filepath.Base(releaseTarballPath)), releaseTarballPath)
			if err != nil {
				return err
			}
			replacedRelease = true
		case "metadata/metadata.yml":
			logger.Printf("\tadding updated tile metadata to new tile")
			mf, err := outZip.Create(file.Name)
			if err != nil {
				return err
			}
			_, err = mf.Write(updatedMetadata)
			if err != nil {
				return err
			}
		default:
			logger.Printf("\tcopying %s from input tile", file.Name)
			err = copyFileToZip(outZip, file)
			if err != nil {
				return err
			}
		}
	}
	if !replacedRelease {
		return fmt.Errorf("failed to replace release %s expected to find a release with path %s", releaseIDFromMetadata(oldRelease), oldRelease.File)
	}
	return outZip.Flush()
}

func addFileToZip(z *zip.Writer, zipPath, filePath string) error {
	// TODO: make this part set header fields to known (maybe constant) values. Such constants would allow us to document expected "expected checksums".
	nf, err := z.Create(zipPath)
	if err != nil {
		return err
	}
	ef, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer closeAndIgnoreError(ef)
	_, err = io.Copy(nf, ef)
	if err != nil {
		return err
	}
	return nil
}

func copyFileToZip(out *zip.Writer, file *zip.File) error {
	nf, err := out.Create(file.Name)
	if err != nil {
		return err
	}
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer closeAndIgnoreError(rc)
	_, err = io.Copy(nf, rc)
	if err != nil {
		return err
	}
	return nil
}

func createReleaseMetadataFromReleaseTarball(filePath string) (boshReleaseMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return boshReleaseMetadata{}, err
	}
	defer closeAndIgnoreError(f)
	mfBuffer, err := readReleaseManifest(f)
	if err != nil {
		return boshReleaseMetadata{}, err
	}
	var manifest struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}
	err = yaml.Unmarshal(mfBuffer, &manifest)
	if err != nil {
		return boshReleaseMetadata{}, err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return boshReleaseMetadata{}, err
	}
	s := sha1.New()
	_, err = io.Copy(s, f)
	if err != nil {
		return boshReleaseMetadata{}, err
	}
	return boshReleaseMetadata{
		Name:    manifest.Name,
		Version: manifest.Version,
		File:    filepath.Base(filePath),
		SHA1:    hex.EncodeToString(s.Sum(nil)),
	}, nil
}

func releaseIDFromMetadata(release boshReleaseMetadata) string {
	return fmt.Sprintf("%s/%s:%s", release.Name, release.Version, release.SHA1)
}

func queryReleaseMetadataFromTile(tilePath, releaseID string) (boshReleaseMetadata, error) {
	metadataBuffer, err := readMetadataFromFile(tilePath)
	if err != nil {
		return boshReleaseMetadata{}, err
	}

	var metadata struct {
		Releases []boshReleaseMetadata `yaml:"releases"`
	}

	err = yaml.Unmarshal(metadataBuffer, &metadata)
	if err != nil {
		return boshReleaseMetadata{}, err
	}

	name, version, digest, err := getNameVersionDigest(releaseID)
	if err != nil {
		return boshReleaseMetadata{}, err
	}

	for _, rel := range metadata.Releases {
		if equalReleaseNameVersionDigest(rel, name, version, digest) {
			return rel, nil
		}
	}

	return boshReleaseMetadata{}, fmt.Errorf("failed to find release in tile equal to %s/%s:%s", name, version, digest)
}

type boshReleaseMetadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	File    string `yaml:"file"`

	SHA1 string `yaml:"sha1"`
}

func equalReleaseNameVersionDigest(rel boshReleaseMetadata, name, version, digest string) bool {
	return rel.Name == name &&
		strings.TrimPrefix(rel.Version, "v") == strings.TrimPrefix(version, "v") &&
		(digest == "" || digest == rel.SHA1)
}

func getNameVersionDigest(s string) (name, version, digest string, _ error) {
	indexOf := func(list []string, target string) int {
		for i, elem := range list {
			if elem == target {
				return i
			}
		}
		return -1
	}
	exp := regexp.MustCompile(`^(?P<name>[^/\n]*)/(?P<version>[^:\n]*)(:(?P<digest>.*))?$`)
	subExNames := exp.SubexpNames()
	nameIndex := indexOf(subExNames, "name")
	versionIndex := indexOf(subExNames, "version")
	digestIndex := indexOf(subExNames, "digest")

	matches := exp.FindStringSubmatch(s)
	if len(matches) < versionIndex {
		return "", "", "", fmt.Errorf("failed to parse release identifier like `%s` in %q", exp.String(), s)
	}

	name = matches[nameIndex]
	version = matches[versionIndex]
	if len(matches) > digestIndex {
		digest = matches[digestIndex]
	}
	return
}

// readReleaseManifest reads from the tarball and parses out the manifest
//
// TODO: remove this if we move the internal/component package to pkg/component in kiln
func readReleaseManifest(releaseTarball io.Reader) ([]byte, error) {
	const releaseManifestFileName = "release.MF"
	zipReader, err := gzip.NewReader(releaseTarball)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(zipReader)

	for {
		h, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if path.Base(h.Name) != releaseManifestFileName {
			continue
		}
		return io.ReadAll(tarReader)
	}

	return nil, fmt.Errorf("%q not found", releaseManifestFileName)
}

func closeAndIgnoreError(c io.Closer) {
	_ = c.Close()
}

func iterateMap(n *yaml.Node, fn func(key, value *yaml.Node)) {
	if n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i]
		value := n.Content[i+1]
		fn(key, value)
	}
}

func displayDiff(logger logger, initialMetadata, finalMetadata string) error {
	_, err := exec.LookPath("diff")
	if err != nil {
		return nil
	}

	initial, err := os.CreateTemp("", "metadata_initial_*.yml")
	_, err = initial.WriteString(initialMetadata)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = initial.Close()
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll(initial.Name())
		if err != nil {
			panic(err)
		}
	}()
	final, err := os.CreateTemp("", "metadata_final_*.yml")
	_, err = final.WriteString(finalMetadata)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = final.Close()
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll(final.Name())
		if err != nil {
			panic(err)
		}
	}()

	stdout, err := exec.Command("diff", "--unified", initial.Name(), final.Name()).Output()
	logger.Print(string(stdout))
	return err
}

func readMetadataFromFile(tilePath string) ([]byte, error) {
	fi, err := os.Stat(tilePath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(tilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return readMetadataFromZip(f, fi.Size())
}

func readMetadataFromZip(ra io.ReaderAt, zipFileSize int64) ([]byte, error) {
	zr, err := zip.NewReader(ra, zipFileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to do open metadata zip reader: %w", err)
	}
	return readMetadataFromFS(zr)
}

func readMetadataFromFS(dir fs.FS) ([]byte, error) {
	metadataFile, err := dir.Open("metadata/metadata.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to do open metadata zip file: %w", err)
	}
	defer func() {
		_ = metadataFile.Close()
	}()
	buf, err := io.ReadAll(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed read metadata: %w", err)
	}
	return buf, nil
}
