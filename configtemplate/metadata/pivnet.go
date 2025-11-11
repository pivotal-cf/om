package metadata

import (
	"github.com/pivotal-cf/om/download_clients"
	"log"
	"os"
)

func NewPivnetProvider(host, token, slug, version, glob string, skipSSL bool) Provider {
	stderr := log.New(os.Stderr, "", 0)
	stdout := log.New(os.Stdout, "", 0)

	proxyConfig := download_clients.ProxyAuthConfig{
		URL:      "", // proxyURL
		Username: "", // proxyUsername
		Password: "", // proxyPassword
		Domain:   "", // proxyDomain
	}
	downloadClient := download_clients.NewPivnetClientWithProxyConfig(
		stdout,
		stderr,
		download_clients.DefaultPivnetFactory,
		token,
		skipSSL,
		host,
		proxyConfig,
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
