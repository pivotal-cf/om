package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

type MetadataExtractor struct{}

type Metadata struct {
	Name    string
	Version string `yaml:"product_version"`
	Raw     []byte
}

func (me MetadataExtractor) ExtractMetadata(productPath string) (Metadata, error) {
	zipReader, err := zip.OpenReader(productPath)
	if err != nil {
		return Metadata{}, err
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		metadataRegexp := regexp.MustCompile(`^(\.\/)?metadata/.*\.yml`)
		matched := metadataRegexp.MatchString(file.Name)

		if matched {
			metadataFile, err := file.Open()
			if err != nil {
				return Metadata{}, err
			}

			contents, err := ioutil.ReadAll(metadataFile)
			if err != nil {
				return Metadata{}, err
			}

			metadata := Metadata{Raw: contents}
			err = yaml.Unmarshal(contents, &metadata)
			if err != nil {
				return Metadata{}, fmt.Errorf("could not extract product metadata: %s", err)
			}

			if metadata.Name == "" || metadata.Version == "" {
				return Metadata{}, errors.New("could not extract product metadata: could not find product details in metadata file")
			}

			return metadata, nil
		}
	}

	return Metadata{}, errors.New("no metadata file was found in provided .pivotal")
}
