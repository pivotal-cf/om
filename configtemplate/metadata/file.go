package metadata

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"regexp"
)

func NewFileProvider(pathToPivotalFile string) Provider {
	return &FileProvider{
		pathToPivotalFile: pathToPivotalFile,
	}
}

type FileProvider struct {
	pathToPivotalFile string
}

var metadataRegexp = regexp.MustCompile("metadata/.*\\.yml")

func (f *FileProvider) MetadataBytes() ([]byte, error) {
	zipReader, err := zip.OpenReader(f.pathToPivotalFile)
	if err != nil {
		return nil, err
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		matched := metadataRegexp.MatchString(file.Name)

		if matched {
			metadataFile, err := file.Open()
			contents, err := ioutil.ReadAll(metadataFile)
			if err != nil {
				return nil, err
			}
			return contents, nil
		}
	}
	return nil, errors.New("no metadata file was found in provided .pivotal")
}
