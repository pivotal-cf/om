package metadata

import (
	"log"
	"os"

	"github.com/pivotal-cf/om/download_clients"
)

func NewPivnetProvider(host, token, slug, version, glob string, skipSSL bool, proxyURL, proxyUsername, proxyPassword, proxyAuthType, proxyKrb5Config string) (Provider, error) {
	stderr := log.New(os.Stderr, "", 0)
	stdout := log.New(os.Stdout, "", 0)

	downloadClient, err := download_clients.NewPivnetClient(
		stdout,
		stderr,
		download_clients.DefaultPivnetFactory,
		token,
		skipSSL,
		host,
		proxyURL,
		proxyUsername,
		proxyPassword,
		proxyAuthType,
		proxyKrb5Config,
	)
	if err != nil {
		return nil, err
	}

	return &PivnetProvider{
		downloadClient: downloadClient,
		slug:           slug,
		version:        version,
		glob:           glob,
		stderr:         stderr,
	}, nil
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
