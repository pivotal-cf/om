package commands

import (
	"errors"
	"fmt"
	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/s3"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


//go:generate counterfeiter -o ./fakes/storer_service.go --fake-name BlobStorer . BlobStorer
type BlobStorer interface {
	ListFiles() ([]string, error)
	DownloadFile(filename string) (io.ReadCloser, error)
}

type S3Configuration struct {
	Bucket string
}

type S3Client struct {
	store BlobStorer
}

func NewS3Client(store BlobStorer) *S3Client {
	return &S3Client{
		store: store,
	}
}

func (s3 S3Client) GetAllProductVersions(slug string) ([]string, error) {
	files, err := s3.store.ListFiles()
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
	return versions, nil

}

func (s3 S3Client) GetLatestProductFile(slug, version, glob string) (*FileArtifact, error) {
	files, err := s3.store.ListFiles()
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
		return nil, fmt.Errorf("the glob '%s' matchs no file", glob)
	}

	return &FileArtifact{Name: artifacts[0]}, nil
}

func (s3 S3Client) DownloadProductToFile(fa *FileArtifact, file *os.File) error {
	f, err := s3.store.DownloadFile(fa.Name)
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, f); err != nil {
		return err
	}

	return nil
}

func (s3 S3Client) DownloadProductStemcell(fa *FileArtifact) (*stemcell, error) {
	return nil, errors.New("downloading stemcells for s3 is not supported at this time")
}

type S3Store struct {
	config S3Configuration
}

func NewS3Store(config S3Configuration) *S3Store {
	return &S3Store{
		config: config,
	}
}

func (s *S3Store) ListFiles() ([]string, error) {
	config := stow.ConfigMap{}
	location, err := stow.Dial("s3", config)
	if err != nil {
		return nil, err
	}
	container, err := location.Container(s.config.Bucket)
	if err != nil {
		return nil, err
	}

	var paths []string
	err = stow.Walk(container, stow.NoPrefix, 100, func(item stow.Item, err error) error {
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

func (s *S3Store) DownloadFile(filename string) (io.ReadCloser, error) {
	config := stow.ConfigMap{}
	location, err := stow.Dial("s3", config)
	if err != nil {
		return nil, err
	}
	container, err := location.Container(s.config.Bucket)
	if err != nil {
		return nil, err
	}
	item, err := container.Item(filename)
	if err != nil {
		return nil, err
	}

	return item.Open()
}