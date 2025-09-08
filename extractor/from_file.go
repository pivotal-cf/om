package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v2"
)

var metadataRegexp = regexp.MustCompile(`^(.*\/)?metadata/.*\.yml`)

func (me *MetadataExtractor) ExtractFromFile(productPath string) (*Metadata, error) {
	zipReader, err := zip.OpenReader(productPath)
	if err != nil {
		return nil, err
	}

	defer zipReader.Close()

	return fromZipFiles(zipReader.File)
}

func fromZipFiles(files []*zip.File) (*Metadata, error) {
	for _, file := range files {
		matched := metadataRegexp.MatchString(file.Name)

		if matched {
			metadata, err := captureMetadata(file)
			if err != nil {
				return nil, err
			}

			return metadata, nil
		}
	}

	return nil, errors.New("no metadata file was found in provided .pivotal")
}

func captureMetadata(file *zip.File) (*Metadata, error) {
	metadataFile, err := file.Open()
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(metadataFile)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{Raw: contents}
	err = yaml.Unmarshal(contents, &metadata)
	if err != nil {
		return nil, fmt.Errorf("could not extract product metadata: %s", err)
	}

	if metadata.Name == "" || metadata.Version == "" {
		return nil, errors.New("could not extract product metadata: could not find product details in metadata file")
	}

	return metadata, nil
}
