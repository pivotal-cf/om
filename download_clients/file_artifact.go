package download_clients

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/om/extractor"
)

type PivnetFileArtifact struct {
	slug        string
	releaseID   int
	productFile pivnet.ProductFile
	client      pivnet.Client
}

func (f PivnetFileArtifact) ProductMetadata() (*extractor.Metadata, error) {
	downloadLink, err := f.productFile.DownloadLink()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve download link: %w", err)
	}

	fetcher := pivnet.NewProductFileLinkFetcher(downloadLink, f.client)
	followedLink, err := fetcher.NewDownloadLink()
	if err != nil {
		return nil, err
	}

	metadata := extractor.NewMetadataExtractor(extractor.WithHTTPClient(f.client.HTTP))
	return metadata.ExtractFromURL(followedLink)
}

func (f PivnetFileArtifact) Name() string {
	return f.productFile.AWSObjectKey
}

func (f PivnetFileArtifact) SHA256() string {
	return f.productFile.SHA256
}

type stowFileArtifact struct {
	name   string
	sha256 string
	source string
}

func (f stowFileArtifact) ProductMetadata() (*extractor.Metadata, error) {
	return nil, fmt.Errorf("there is no way to extract metadata from source \"%s\"", f.source)
}

func (f stowFileArtifact) Name() string {
	return f.name
}

func (f stowFileArtifact) SHA256() string {
	return f.sha256
}
