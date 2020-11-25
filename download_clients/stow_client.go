package download_clients

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/graymeta/stow"
	"github.com/pivotal-cf/om/extractor"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ./fakes/config_service.go --fake-name Config . StowConfiger
type StowConfiger interface {
	Config(name string) (string, bool)
	Set(name, value string)
}

type Stower interface {
	Dial(kind string, config StowConfiger) (stow.Location, error)
	Walk(container stow.Container, prefix string, pageSize int, fn stow.WalkFunc) error
}

type StowWrapper struct{}

func (d StowWrapper) Dial(kind string, config StowConfiger) (stow.Location, error) {
	location, err := stow.Dial(kind, config)
	return location, err
}
func (d StowWrapper) Walk(container stow.Container, prefix string, pageSize int, fn stow.WalkFunc) error {
	return stow.Walk(container, prefix, pageSize, fn)
}

type stowClient struct {
	stower       Stower
	bucket       string
	Config       stow.Config
	productPath  string
	stemcellPath string
	kind         string
	stderr       *log.Logger
}

func NewStowClient(stower Stower, stderr *log.Logger, config stow.ConfigMap, productPath string, stemcellPath string, kind string, bucket string) stowClient {
	return stowClient{
		stower:       stower,
		bucket:       bucket,
		Config:       config,
		productPath:  productPath,
		stemcellPath: stemcellPath,
		kind:         kind,
		stderr:       stderr,
	}
}

func (s stowClient) Name() string {
	return s.kind
}

func (s stowClient) GetAllProductVersions(slug string) ([]string, error) {
	return s.getAllProductVersionsFromPath(slug, s.productPath)
}

func (s stowClient) getAllProductVersionsFromPath(slug, path string) ([]string, error) {
	files, err := s.listFiles()
	if err != nil {
		return nil, err
	}

	productFileCompiledRegex := regexp.MustCompile(
		fmt.Sprintf(`^/?%s/?\[%s,(.*?)\]`,
			regexp.QuoteMeta(strings.Trim(path, "/")),
			slug,
		),
	)

	var versions []string
	versionFound := make(map[string]bool)
	for _, fileName := range files {
		match := productFileCompiledRegex.FindStringSubmatch(fileName)
		if match != nil {
			version := match[1]
			if !versionFound[version] {
				versions = append(versions, version)
				versionFound[version] = true
			}
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no files matching pivnet-product-slug %s found", slug)
	}

	return versions, nil
}

func (s *stowClient) listFiles() ([]string, error) {
	container, err := s.getContainer()
	if err != nil {
		return nil, err
	}

	var paths []string
	err = s.stower.Walk(container, stow.NoPrefix, 100, func(item stow.Item, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, item.ID())
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("bucket '%s' contains no files", s.bucket)
	}

	return paths, nil
}

func (s *stowClient) getContainer() (stow.Container, error) {
	location, err := s.stower.Dial(s.kind, s.Config)
	if err != nil {
		return nil, err
	}
	container, err := location.Container(s.bucket)
	if err != nil {
		endpoint, _ := s.Config.Config("endpoint")
		if endpoint != "" {
			return nil, fmt.Errorf(
				"could not reach provided endpoint and bucket '%s/%s': %s\nCheck bucket and endpoint configuration",
				endpoint,
				s.bucket,
				err,
			)
		}
		return nil, fmt.Errorf(
			"could not reach provided bucket '%s': %s\nCheck bucket and endpoint configuration",
			s.bucket,
			err,
		)
	}
	return container, nil
}

func (s stowClient) GetLatestProductFile(slug, version, glob string) (FileArtifacter, error) {
	files, err := s.listFiles()
	if err != nil {
		return nil, err
	}

	validFile := regexp.MustCompile(
		fmt.Sprintf(`^/?(%s|%s)/?\[%s,%s\]`,
			regexp.QuoteMeta(strings.Trim(s.productPath, "/")),
			regexp.QuoteMeta(strings.Trim(s.stemcellPath, "/")),
			slug,
			regexp.QuoteMeta(version),
		),
	)
	var prefixedFilepaths []string
	var globMatchedFilepaths []string

	for _, f := range files {
		if validFile.MatchString(f) {
			prefixedFilepaths = append(prefixedFilepaths, f)
		}
	}

	if len(prefixedFilepaths) == 0 {
		return nil, fmt.Errorf("no product files with expected prefix [%s,%s] found. Please ensure the file you're trying to download was initially persisted from Pivotal Network net using an appropriately configured download-product command", slug, version)
	}

	for _, f := range prefixedFilepaths {
		removePrefixRegex := regexp.MustCompile(`^\[.*\]`)
		baseFilename := removePrefixRegex.ReplaceAllString(filepath.Base(f), "")

		matched, _ := filepath.Match(glob, baseFilename)
		if matched {
			globMatchedFilepaths = append(globMatchedFilepaths, f)
		}
	}

	if len(globMatchedFilepaths) > 1 {
		return nil, fmt.Errorf("the glob '%s' matches multiple files. Write your glob to match exactly one of the following:\n  %s", glob, strings.Join(globMatchedFilepaths, "\n  "))
	}

	if len(globMatchedFilepaths) == 0 {
		availableFiles := strings.Join(prefixedFilepaths, ", ")
		if availableFiles == "" {
			availableFiles = "none"
		}
		return nil, fmt.Errorf("the glob '%s' matches no file\navailable files: %s", glob, availableFiles)
	}

	return &stowFileArtifact{name: globMatchedFilepaths[0], source: s.kind}, nil
}

func (s stowClient) DownloadProductToFile(fa FileArtifacter, destinationFile *os.File) error {
	blobReader, size, err := s.initializeBlobReader(fa.Name())
	if err != nil {
		return err
	}

	progressBar, wrappedBlobReader := s.startProgressBar(size, blobReader)
	defer progressBar.Finish()

	if err = s.streamBufferToFile(destinationFile, wrappedBlobReader); err != nil {
		return err
	}

	return nil
}

func (s *stowClient) initializeBlobReader(filename string) (blobToRead io.ReadCloser, fileSize int64, err error) {
	container, err := s.getContainer()
	if err != nil {
		return nil, 0, err
	}

	item, err := container.Item(filename)
	if err != nil {
		return nil, 0, err
	}

	fileSize, err = item.Size()
	if err != nil {
		return nil, 0, err
	}
	blobToRead, err = item.Open()
	return blobToRead, fileSize, err
}

func (s stowClient) startProgressBar(size int64, item io.Reader) (*pb.ProgressBar, io.Reader) {
	progressBar := pb.Default.New(0)
	progressBar.SetWriter(s.stderr.Writer())
	progressBar.Set(pb.Bytes, true)
	progressBar.SetTotal(size)
	progressBar.SetMaxWidth(80)
	reader := progressBar.NewProxyReader(item)
	progressBar.Start()
	return progressBar, reader
}

func (s stowClient) streamBufferToFile(destinationFile *os.File, wrappedBlobReader io.Reader) error {
	_, err := io.Copy(destinationFile, wrappedBlobReader)
	return err
}

func (s stowClient) GetLatestStemcellForProduct(_ FileArtifacter, downloadedProductFileName string) (StemcellArtifacter, error) {
	definedStemcell, err := stemcellFromProduct(downloadedProductFileName)
	if err != nil {
		return nil, err
	}

	definedMajor, definedPatch, err := stemcellVersionPartsFromString(definedStemcell.Version())
	if err != nil {
		return nil, err
	}

	allStemcellVersions, err := s.getAllProductVersionsFromPath(definedStemcell.Slug(), s.stemcellPath)
	if err != nil {
		return nil, fmt.Errorf("could not find stemcells on %s: %s", s.kind, err)
	}

	var filteredVersions []string
	for _, version := range allStemcellVersions {
		major, patch, _ := stemcellVersionPartsFromString(version)

		if major == definedMajor && patch >= definedPatch {
			filteredVersions = append(filteredVersions, version)
		}
	}

	if len(filteredVersions) == 0 {
		return nil, fmt.Errorf("no versions could be found equal to or greater than %s", definedStemcell.Version())
	}

	latestVersion, err := getLatestStemcellVersion(filteredVersions)
	if err != nil {
		return nil, err
	}

	return &stemcell{
		version: latestVersion,
		slug:    definedStemcell.Slug(),
	}, nil
}

func stemcellFromProduct(filename string) (*stemcell, error) {
	metadataExtractor := extractor.MetadataExtractor{}
	metadata, err := metadataExtractor.ExtractFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not find the appropriate stemcell associated with the tile %q: %s", filename, err)
	}

	stemcellNameToPivnetProductName := map[string]string{
		"ubuntu-xenial": "stemcells-ubuntu-xenial",
		"ubuntu-trusty": "stemcells",
		"windows2016":   "stemcells-windows-server",
		"windows1803":   "stemcells-windows-server",
		"windows2019":   "stemcells-windows-server",
	}

	return &stemcell{
		slug:    stemcellNameToPivnetProductName[metadata.StemcellCriteria.OS],
		version: metadata.StemcellCriteria.Version,
	}, nil
}
