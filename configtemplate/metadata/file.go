package metadata

import (
	"fmt"
	"github.com/pivotal-cf/om/extractor"
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
	metadataExtractor := extractor.MetadataExtractor{}
	metadata, err := metadataExtractor.ExtractMetadata(f.pathToPivotalFile)
	if err != nil {
		return nil, fmt.Errorf("could not extract metadata from %q: %s", f.pathToPivotalFile, err)
	}

	return metadata.Raw, nil
}
