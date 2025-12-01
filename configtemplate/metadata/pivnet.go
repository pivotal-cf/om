package metadata

import (
	"github.com/pivotal-cf/om/download_clients"
	"log"
	"os"
)

func NewPivnetProvider(host, token, slug, version, glob string, skipSSL bool) Provider {
	stderr := log.New(os.Stderr, "", 0)
	stdout := log.New(os.Stdout, "", 0)

	downloadClient := download_clients.NewPivnetClient(
		stdout,
		stderr,
		download_clients.DefaultPivnetFactory,
		token,
		skipSSL,
		host,
		"", // proxyURL
		"", // proxyUsername
		"", // proxyPassword
		"", // proxyAuthType
		"", // proxyKrb5Config
	)

	return &PivnetProvider{
		downloadClient: downloadClient,
		slug:           slug,
		version:        version,
		glob:           glob,
		stderr:         stderr,
	}
}

type PivnetProvider struct {
	downloadClient download_clients.ProductDownloader
	slug           string
	version        string
	glob           string
	stderr         *log.Logger
}

func (p *PivnetProvider) MetadataBytes() ([]byte, error) {
	fileArtifact, err := p.downloadClient.GetLatestProductFile(p.slug, p.version, p.glob)
	if err != nil {
		return nil, err
	}

	metadata, err := fileArtifact.ProductMetadata()
	if err != nil {
		return nil, err
	}
	return metadata.Raw, nil
}
