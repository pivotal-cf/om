package download_clients

import (
	"archive/zip"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"github.com/pivotal-cf/om/progress"
	"gopkg.in/go-playground/validator.v9"
)

//go:generate counterfeiter -o ./fakes/config_service.go --fake-name Config . Config
type Config interface {
	Config(name string) (string, bool)
	Set(name, value string)
}

type Stower interface {
	Dial(kind string, config Config) (stow.Location, error)
	Walk(container stow.Container, prefix string, pageSize int, fn stow.WalkFunc) error
}

type S3Configuration struct {
	Bucket          string `validate:"required"`
	AccessKeyID     string
	SecretAccessKey string
	RegionName      string `validate:"required"`
	Endpoint        string
	DisableSSL      bool
	EnableV2Signing bool
	ProductPath     string
	StemcellPath    string
	AuthType        string
}

type S3Client struct {
	stower         Stower
	bucket         string
	Config         stow.Config
	progressWriter io.Writer
	productPath    string
	stemcellPath   string
}

func NewS3Client(stower Stower, config S3Configuration, progressWriter io.Writer) (*S3Client, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return nil, err
	}

	disableSSL := strconv.FormatBool(config.DisableSSL)
	enableV2Signing := strconv.FormatBool(config.EnableV2Signing)
	if config.AuthType == "" {
		config.AuthType = "accesskey"
	}

	err = validateAccessKeyAuthType(config)
	if err != nil {
		return nil, err
	}

	stowConfig := stow.ConfigMap{
		s3.ConfigAccessKeyID: config.AccessKeyID,
		s3.ConfigSecretKey:   config.SecretAccessKey,
		s3.ConfigRegion:      config.RegionName,
		s3.ConfigEndpoint:    config.Endpoint,
		s3.ConfigDisableSSL:  disableSSL,
		s3.ConfigV2Signing:   enableV2Signing,
		s3.ConfigAuthType:    config.AuthType,
	}

	return &S3Client{
		stower:         stower,
		Config:         stowConfig,
		bucket:         config.Bucket,
		progressWriter: progressWriter,
		productPath:    config.ProductPath,
		stemcellPath:   config.StemcellPath,
	}, nil
}

func (s3 S3Client) GetAllProductVersions(slug string) ([]string, error) {
	return s3.getAllProductVersionsFromPath(slug, s3.productPath)
}

func (s3 S3Client) getAllProductVersionsFromPath(slug, path string) ([]string, error) {
	files, err := s3.listFiles()
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

func (s3 S3Client) GetLatestProductFile(slug, version, glob string) (*FileArtifact, error) {
	files, err := s3.listFiles()
	if err != nil {
		return nil, err
	}

	validFile := regexp.MustCompile(
		fmt.Sprintf(`^/?(%s|%s)/?\[%s,%s\]`,
			regexp.QuoteMeta(strings.Trim(s3.productPath, "/")),
			regexp.QuoteMeta(strings.Trim(s3.stemcellPath, "/")),
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

	return &FileArtifact{Name: globMatchedFilepaths[0]}, nil
}

func (s3 S3Client) DownloadProductToFile(fa *FileArtifact, destinationFile *os.File) error {
	blobReader, size, err := s3.initializeBlobReader(fa.Name)
	if err != nil {
		return err
	}

	progressBar, wrappedBlobReader := s3.startProgressBar(size, blobReader)
	defer progressBar.Finish()

	if err = s3.streamBufferToFile(destinationFile, wrappedBlobReader); err != nil {
		return err
	}

	return nil
}

func (s3 *S3Client) initializeBlobReader(filename string) (blobToRead io.ReadCloser, fileSize int64, err error) {
	container, err := s3.getContainer()
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

func (s3 S3Client) startProgressBar(size int64, item io.Reader) (progressBar *progress.Bar, reader io.Reader) {
	progressBar = progress.NewBar()
	progressBar.SetTotal64(size)
	progressBar.SetOutput(s3.progressWriter)
	reader = progressBar.NewProxyReader(item)
	_, _ = s3.progressWriter.Write([]byte("Downloading product from s3..."))
	progressBar.Start()
	return progressBar, reader
}

func (s3 S3Client) streamBufferToFile(destinationFile *os.File, wrappedBlobReader io.Reader) error {
	_, err := io.Copy(destinationFile, wrappedBlobReader)
	return err
}

func (s3 S3Client) GetLatestStemcellForProduct(_ *FileArtifact, downloadedProductFileName string) (*Stemcell, error) {
	definedStemcell, err := stemcellFromProduct(downloadedProductFileName)
	if err != nil {
		return nil, err
	}

	definedMajor, definedPatch, err := stemcellVersionPartsFromString(definedStemcell.Version)
	if err != nil {
		return nil, err
	}

	allStemcellVersions, err := s3.getAllProductVersionsFromPath(definedStemcell.Slug, s3.stemcellPath)
	if err != nil {
		return nil, fmt.Errorf("could not find stemcells on s3: %s", err)
	}

	var filteredVersions []string
	for _, version := range allStemcellVersions {
		major, patch, _ := stemcellVersionPartsFromString(version)

		if major == definedMajor && patch >= definedPatch {
			filteredVersions = append(filteredVersions, version)
		}
	}

	if len(filteredVersions) == 0 {
		return nil, fmt.Errorf("no versions could be found equal to or greater than %s", definedStemcell.Version)
	}

	latestVersion, err := getLatestStemcellVersion(filteredVersions)
	if err != nil {
		return nil, err
	}

	return &Stemcell{
		Version: latestVersion,
		Slug:    definedStemcell.Slug,
	}, nil
}

func (s3 *S3Client) listFiles() ([]string, error) {
	container, err := s3.getContainer()
	if err != nil {
		return nil, err
	}

	var paths []string
	err = s3.stower.Walk(container, stow.NoPrefix, 100, func(item stow.Item, err error) error {
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
		return nil, fmt.Errorf("bucket contains no files")
	}

	return paths, nil
}

func (s3 *S3Client) getContainer() (stow.Container, error) {
	location, err := s3.stower.Dial("s3", s3.Config)
	if err != nil {
		return nil, err
	}
	container, err := location.Container(s3.bucket)
	if err != nil {
		endpoint, _ := s3.Config.Config("endpoint")
		if endpoint != "" {
			return nil, fmt.Errorf(
				"could not reach provided endpoint and bucket '%s/%s': %s\nCheck bucket and endpoint configuration",
				endpoint,
				s3.bucket,
				err,
			)
		}
		return nil, fmt.Errorf(
			"could not reach provided bucket '%s': %s\nCheck bucket and endpoint configuration",
			s3.bucket,
			err,
		)
	}
	return container, nil
}

func validateAccessKeyAuthType(config S3Configuration) error {
	if config.AuthType == "iam" {
		return nil
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return fmt.Errorf("the flags \"s3-access-key-id\" and \"s3-secret-access-key\" are required when the \"auth-type\" is \"accesskey\"")
	}

	return nil
}

type stemcellMetadata struct {
	Metadata internalStemcellMetadata `yaml:"stemcell_criteria"`
}

type internalStemcellMetadata struct {
	Os                   string `yaml:"os"`
	Version              string `yaml:"version"`
	PatchSecurityUpdates string `yaml:"enable_patch_security_updates"`
}

func stemcellFromProduct(filename string) (*Stemcell, error) {
	// Open a zip archive for reading.
	tileZipReader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("could not parse tile. Ensure that downloaded file is a valid pivotal tile: %s", err)
	}

	defer tileZipReader.Close()

	metadataRegex := regexp.MustCompile(`^metadata/.*\.yml`)

	for _, file := range tileZipReader.File {
		// check if the file matches the name for application portfolio xml

		if metadataRegex.MatchString(file.Name) {
			metadataReadCloser, err := file.Open()
			if err != nil {
				return nil, err
			}

			metadataBuffer := new(bytes.Buffer)
			_, err = metadataBuffer.ReadFrom(metadataReadCloser)
			if err != nil {
				return nil, err
			}

			metadata := stemcellMetadata{}
			err = yaml.Unmarshal(metadataBuffer.Bytes(), &metadata)
			if err != nil {
				return nil, err
			}

			stemcellNameToPivnetProductName := map[string]string{
				"ubuntu-xenial": "stemcells-ubuntu-xenial",
				"ubuntu-trusty": "stemcells",
				"windows2016":   "stemcells-windows-server",
				"windows1803":   "stemcells-windows-server",
				"windows2019":   "stemcells-windows-server",
			}

			return &Stemcell{
				Slug:    stemcellNameToPivnetProductName[metadata.Metadata.Os],
				Version: metadata.Metadata.Version,
			}, nil
		}
	}
	return nil, fmt.Errorf("could not find the appropriate stemcell associated with the tile: %s", filename)
}
