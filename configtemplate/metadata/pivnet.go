package metadata

import (
	"fmt"
	"github.com/pivotal-cf/om/extractor"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	pivnetapi "github.com/pivotal-cf/go-pivnet/v5"
	"github.com/pivotal-cf/go-pivnet/v5/logshim"
)

func NewPivnetProvider(host, token, slug, version, glob string, skipSSL bool) Provider {
	logWriter := os.Stderr
	logger := log.New(logWriter, "", log.LstdFlags)
	config := pivnetapi.ClientConfig{
		Host:              host,
		UserAgent:         "tile-config-generator",
		SkipSSLValidation: skipSSL,
	}
	ts := pivnetapi.NewAccessTokenOrLegacyToken(token, config.Host, skipSSL, config.UserAgent)
	ls := logshim.NewLogShim(logger, logger, false)
	client := pivnetapi.NewClient(ts, config, ls)

	return &PivnetProvider{
		client:         client,
		progressWriter: os.Stderr,
		logger:         ls,
		slug:           slug,
		version:        version,
		glob:           glob,
	}
}

type PivnetProvider struct {
	client         pivnetapi.Client
	progressWriter io.Writer
	logger         *logshim.LogShim
	slug           string
	version        string
	glob           string
}

func (p *PivnetProvider) MetadataBytes() ([]byte, error) {
	releases, err := p.client.Releases.List(p.slug)
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if release.Version == p.version {
			productFiles, err := p.client.ProductFiles.ListForRelease(p.slug, release.ID)
			if err != nil {
				return nil, err
			}

			return p.downloadFiles(productFiles, release.ID)
		}
	}

	var list []string
	for _, release := range releases {
		list = append(list, release.Version)
	}

	return nil, fmt.Errorf("no version matched for slug %s, version %s and glob %s.\nVersions found:\n  %s", p.slug, p.version, p.glob, strings.Join(list, "\n  "))
}

func (p *PivnetProvider) downloadFiles(
	productFiles []pivnetapi.ProductFile,
	releaseID int,
) ([]byte, error) {
	filtered, err := productFileKeysByGlobs(productFiles, p.glob)
	if err != nil {
		return nil, err
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no file matched for slug %s, releaseID %d and glob %s", p.slug, releaseID, p.glob)
	}
	if len(filtered) > 1 {
		list := []string{}
		for _, file := range filtered {
			list = append(list, path.Base(file.AWSObjectKey))
		}
		return nil, fmt.Errorf("the glob '%s' matches multiple files. Write your glob to match exactly one of the following:\n  %s", p.glob, strings.Join(list, "\n  "))
	}

	err = p.client.EULA.Accept(p.slug, releaseID)
	if err != nil {
		return nil, err
	}

	pf := filtered[0]
	downloadLink, err := pf.DownloadLink()
	if err != nil {
		return nil, err
	}

	fetcher := pivnetapi.NewProductFileLinkFetcher(downloadLink, p.client)
	followedLink, err := fetcher.NewDownloadLink()
	if err != nil {
		return nil, err
	}

	metadataExtractor := extractor.NewMetadataExtractor(extractor.WithHTTPClient(p.client.HTTP))
	metadata, err := metadataExtractor.ExtractFromURL(followedLink)
	if err != nil {
		return nil, fmt.Errorf("could not extract metadata from %q: %s", followedLink, err)
	}

	return metadata.Raw, nil
}

func productFileKeysByGlobs(
	productFiles []pivnetapi.ProductFile,
	pattern string,
) ([]pivnetapi.ProductFile, error) {
	filtered := []pivnetapi.ProductFile{}

	for _, p := range productFiles {
		parts := strings.Split(p.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		matched, err := filepath.Match(pattern, fileName)
		if err != nil {
			return nil, err
		}

		if matched {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 && pattern != "" {
		return nil, fmt.Errorf("no match for pattern: '%s'", pattern)
	}

	return filtered, nil
}
