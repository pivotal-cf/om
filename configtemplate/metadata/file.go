package metadata

import (
	"archive/zip"
)

func NewFileProvider(pathToPivotalFile string) Provider {
	return &FileProvider{
		pathToPivotalFile: pathToPivotalFile,
	}
}

type FileProvider struct {
	pathToPivotalFile string
}

func (f *FileProvider) MetadataBytes() ([]byte, error) {
	zipReader, err := zip.OpenReader(f.pathToPivotalFile)
	if err != nil {
		return nil, err
	}

	defer zipReader.Close()

	return ExtractMetadataFromZip(&zipReader.Reader)
}
