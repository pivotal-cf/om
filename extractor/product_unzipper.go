package extractor

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type ProductUnzipper struct{}

func (u ProductUnzipper) ExtractMetadata(productPath string) (string, string, error) {
	zipReader, err := zip.OpenReader(productPath)
	if err != nil {
		return "", "", err
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, ".yml") {
			metadataFile, err := file.Open()
			if err != nil {
				return "", "", err
			}

			contents, err := ioutil.ReadAll(metadataFile)
			if err != nil {
				return "", "", err
			}

			var metadata struct {
				Name           string
				ProductVersion string `yaml:"product_version"`
			}
			err = yaml.Unmarshal(contents, &metadata)
			if err != nil {
				return "", "", fmt.Errorf("could not extract product metadata: %s", err)
			}

			if metadata.Name == "" || metadata.ProductVersion == "" {
				return "", "", errors.New("could not extract product metadata: could not find product details in metadata file")
			}

			return metadata.Name, metadata.ProductVersion, nil
		}
	}

	return "", "", errors.New("no metadata file was found in provided .pivotal")
}
