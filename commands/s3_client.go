package commands

import (
	"errors"
	"fmt"
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
	Bucket          string `yaml:"bucket" validate:"required"`
	AccessKeyID     string `yaml:"access-key-id" validate:"required"`
	SecretAccessKey string `yaml:"secret-access-key" validate:"required"`
	RegionName      string `yaml:"region-name"`
	Endpoint        string `yaml:"endpoint"`
	DisableSSL      bool   `yaml:"disable-ssl" `
	EnableV2Signing bool   `yaml:"enable-v2-signing" `
}

type S3Client struct {
	stower         Stower
	bucket         string
	Config         stow.Config
	progressWriter io.Writer
}

func NewS3Client(stower Stower, config S3Configuration, progressWriter io.Writer) (*S3Client, error) {
	validate := validator.New()
	err := validate.Struct(config)
	if err != nil {
		return nil, err
	}

	if config.Endpoint == "" && config.RegionName == "" {
		return nil, fmt.Errorf("no endpoint information provided in config file; please provide either region or endpoint")
	}

	disableSSL := strconv.FormatBool(config.DisableSSL)
	enableV2Signing := strconv.FormatBool(config.EnableV2Signing)
	stowConfig := stow.ConfigMap{
		s3.ConfigAccessKeyID: config.AccessKeyID,
		s3.ConfigSecretKey:   config.SecretAccessKey,
		s3.ConfigRegion:      config.RegionName,
		s3.ConfigEndpoint:    config.Endpoint,
		s3.ConfigDisableSSL:  disableSSL,
		s3.ConfigV2Signing:   enableV2Signing,
	}

	return &S3Client{
		stower:         stower,
		Config:         stowConfig,
		bucket:         config.Bucket,
		progressWriter: progressWriter,
	}, nil
}

func (s3 S3Client) GetAllProductVersions(slug string) ([]string, error) {
	files, err := s3.ListFiles()
	if err != nil {
		return nil, err
	}

	validFile := regexp.MustCompile(fmt.Sprintf(`^%s-(.*?)_`, slug))

	var versions []string
	versionFound := make(map[string]bool)
	for _, f := range files {
		x := validFile.FindStringSubmatch(f)
		if len(x) == 2 {
			version := x[1]
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
	files, err := s3.ListFiles()
	if err != nil {
		return nil, err
	}

	validFile := regexp.MustCompile(fmt.Sprintf(`^%s-%s`, slug, version))
	var artifacts []string

	for _, f := range files {
		if validFile.MatchString(f) {
			matched, _ := filepath.Match(glob, f)
			if matched {
				artifacts = append(artifacts, f)
			}
		}
	}

	if len(artifacts) > 1 {
		return nil, fmt.Errorf("the glob '%s' matches multiple files. Write your glob to match exactly one of the following:\n  %s", glob, strings.Join(artifacts, "\n  "))
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("the glob '%s' matches no file", glob)
	}

	return &FileArtifact{Name: artifacts[0]}, nil
}

func (s3 S3Client) DownloadProductToFile(fa *FileArtifact, file *os.File) error {
	item, size, err := s3.DownloadFile(fa.Name)
	if err != nil {
		return err
	}

	progressBar := progress.NewBar()
	progressBar.SetTotal64(size)
	progressBar.SetOutput(s3.progressWriter)
	reader := progressBar.NewProxyReader(item)
	s3.progressWriter.Write([]byte("Downloading product from s3..."))
	progressBar.Start()
	defer progressBar.Finish()

	if _, err := io.Copy(file, reader); err != nil {
		return err
	}

	return nil
}

func (s3 S3Client) DownloadProductStemcell(fa *FileArtifact) (*stemcell, error) {
	return nil, errors.New("downloading stemcells for s3 is not supported at this time")
}

var invalidSignatureErrorMessage = "could not contact S3 with the endpoint provided. Please validate that the endpoint is a valid S3 endpoint"

/* TODO: this should be a private method */
func (s *S3Client) ListFiles() ([]string, error) {
	location, err := s.stower.Dial("s3", s.Config)
	if err != nil {
		return nil, err
	}
	container, err := location.Container(s.bucket)
	if err != nil {
		if strings.Contains(err.Error(), "<InvalidSignatureException>") {
			return nil, errors.New(invalidSignatureErrorMessage)
		}
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

	return paths, nil
}

/* TODO: this should be a private method in DownloadProductToFile */
func (s *S3Client) DownloadFile(filename string) (io.ReadCloser, int64, error) {
	location, err := s.stower.Dial("s3", s.Config)
	if err != nil {
		return nil, 0, err
	}
	container, err := location.Container(s.bucket)
	if err != nil {
		if strings.Contains(err.Error(), "<InvalidSignatureException>") {
			return nil, 0, errors.New(invalidSignatureErrorMessage)
		}
		return nil, 0, err
	}
	item, err := container.Item(filename)
	if err != nil {
		return nil, 0, err
	}

	size, err := item.Size()
	if err != nil {
		return nil, 0, err
	}
	readcloser, err := item.Open()
	return readcloser, size, err
}
