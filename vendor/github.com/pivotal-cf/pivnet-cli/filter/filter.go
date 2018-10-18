package filter

import (
	"path/filepath"
	"regexp"
	"strings"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
)

type Filter struct {
	l logger.Logger
}

func NewFilter(l logger.Logger) *Filter {
	return &Filter{
		l: l,
	}
}

// ReleasesByVersion returns all releases that match the provided version regex
func (f Filter) ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error) {
	filteredReleases := make([]pivnet.Release, 0)

	for _, release := range releases {
		match, err := regexp.MatchString(version, release.Version)
		if err != nil {
			return nil, err
		}

		if match {
			filteredReleases = append(filteredReleases, release)
		}
	}

	return filteredReleases, nil
}

func (f Filter) ProductFileKeysByGlobs(
	productFiles []pivnet.ProductFile,
	globs []string,
) ([]pivnet.ProductFile, error) {
	f.l.Debug("filter.ProductFilesKeysByGlobs", logger.Data{"globs": globs})

	filtered := []pivnet.ProductFile{}
	for _, p := range productFiles {
		parts := strings.Split(p.AWSObjectKey, "/")
		fileName := parts[len(parts)-1]

		for _, pattern := range globs {
			matched, err := filepath.Match(pattern, fileName)
			if err != nil {
				return nil, err
			}

			if matched {
				filtered = append(filtered, p)
			}
		}
	}

	return filtered, nil
}
