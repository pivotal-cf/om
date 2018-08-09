package baking

import "io/ioutil"

type MetadataService struct {
	metadataPath string
}

func NewMetadataService() MetadataService {
	return MetadataService{}
}

func (ms MetadataService) Read(path string) ([]byte, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
