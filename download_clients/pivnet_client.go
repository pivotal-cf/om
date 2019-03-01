package download_clients

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet"
	pivnetlog "github.com/pivotal-cf/go-pivnet/logger"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

//go:generate counterfeiter -o ./fakes/pivnet_downloader_service.go --fake-name PivnetDownloader . PivnetDownloader
type PivnetDownloader interface {
	ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error)
	ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	DownloadProductFile(location *os.File, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
}

//go:generate counterfeiter -o ./fakes/pivnet_filter_service.go --fake-name PivnetFilter . PivnetFilter
type PivnetFilter interface {
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
	ProductFileKeysByGlobs(productFiles []pivnet.ProductFile, globs []string) ([]pivnet.ProductFile, error)
}

type PivnetFactory func(config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader

func NewPivnetClient(logger pivnetlog.Logger, progressWriter io.Writer, factory PivnetFactory, token string, filter PivnetFilter) *pivnetClient {
	downloader := factory(
		pivnet.ClientConfig{
			Host:              pivnet.DefaultHost,
			Token:             token,
			UserAgent:         fmt.Sprintf("om-download-product"),
			SkipSSLValidation: false,
		},
		logger)

	return &pivnetClient{
		filter:         filter,
		progressWriter: progressWriter,
		downloader:     downloader,
	}
}

type pivnetClient struct {
	downloader     PivnetDownloader
	filter         PivnetFilter
	progressWriter io.Writer
}

type FileArtifact struct {
	Name          string
	SHA256        string
	slug          string
	releaseID     int
	productFileID int
}

type Stemcell struct {
	Slug    string
	Version string
}

func (p *pivnetClient) GetAllProductVersions(slug string) ([]string, error) {
	releases, err := p.downloader.ReleasesForProductSlug(slug)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.Version)
	}
	return versions, nil
}

func (p *pivnetClient) GetLatestProductFile(slug, version, glob string) (*FileArtifact, error) {
	// 1. Check the release for given version / slug
	release, err := p.downloader.ReleaseForVersion(slug, version)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the release for %s %s: %s", slug, version, err)
	}

	// 2. Get filename from pivnet
	productFiles, err := p.downloader.ProductFilesForRelease(slug, release.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the product files for %s %s: %s", slug, version, err)
	}

	productFiles, err = p.filter.ProductFileKeysByGlobs(productFiles, []string{glob})
	if err != nil {
		return nil, fmt.Errorf("could not glob product files: %s", err)
	}

	if err := p.checkForSingleProductFile(glob, productFiles); err != nil {
		return nil, err
	}

	return &FileArtifact{
		Name:          productFiles[0].AWSObjectKey,
		SHA256:        productFiles[0].SHA256,
		releaseID:     release.ID,
		slug:          slug,
		productFileID: productFiles[0].ID,
	}, nil
}

func (p *pivnetClient) DownloadProductToFile(fa *FileArtifact, file *os.File) error {
	err := p.downloader.DownloadProductFile(file, fa.slug, fa.releaseID, fa.productFileID, p.progressWriter)
	if err != nil {
		return fmt.Errorf("could not download product file %s: %s", fa.slug, err)
	}
	return nil
}

func (p *pivnetClient) GetLatestStemcellForProduct(fa *FileArtifact, _ string) (*Stemcell, error) {
	dependencies, err := p.downloader.ReleaseDependencies(fa.slug, fa.releaseID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch stemcell dependency for %s: %s", fa.slug, err)
	}

	stemcellSlug, stemcellVersion, err := p.getLatestStemcell(dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not sort stemcell dependency: %s", err)
	}

	return &Stemcell{Slug: stemcellSlug, Version: stemcellVersion}, nil
}

func (p *pivnetClient) checkForSingleProductFile(glob string, productFiles []pivnet.ProductFile) error {
	if len(productFiles) > 1 {
		var productFileNames []string
		for _, productFile := range productFiles {
			productFileNames = append(productFileNames, path.Base(productFile.AWSObjectKey))
		}
		return fmt.Errorf("the glob '%s' matches multiple files. Write your glob to match exactly one of the following:\n  %s", glob, strings.Join(productFileNames, "\n  "))
	} else if len(productFiles) == 0 {
		return fmt.Errorf("the glob '%s' matches no file", glob)
	}

	return nil
}

func (p *pivnetClient) getLatestStemcell(dependencies []pivnet.ReleaseDependency) (string, string, error) {
	var (
		stemcellSlug string
		versions     []string
	)

	for _, dependency := range dependencies {
		if strings.Contains(dependency.Release.Product.Slug, "stemcells") {
			stemcellSlug = dependency.Release.Product.Slug
			versions = append(versions, dependency.Release.Version)
		}
	}

	stemcellVersion, err := getLatestStemcellVersion(versions)
	if err != nil {
		return "", "", err
	}

	return stemcellSlug, stemcellVersion, nil
}

const errorForVersion = "versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: %s"

func getLatestStemcellVersion(versions []string) (string, error) {
	var (
		stemcellVersion      string
		stemcellVersionMajor int
		stemcellVersionMinor int
	)

	for _, versionString := range versions {
		major, minor, err := stemcellVersionPartsFromString(versionString)
		if err != nil {
			return "", err
		}

		if major > stemcellVersionMajor {
			stemcellVersionMajor = major
			stemcellVersionMinor = minor
			stemcellVersion = versionString
		} else if major == stemcellVersionMajor && minor > stemcellVersionMinor {
			stemcellVersionMinor = minor
			stemcellVersion = versionString
		}
	}

	return stemcellVersion, nil
}

func stemcellVersionPartsFromString(version string) (int, int, error) {
	splitVersions := strings.Split(version, ".")
	if len(splitVersions) == 1 {
		splitVersions = []string{splitVersions[0], "0"}
	}
	if len(splitVersions) != 2 {
		return 0, 0, fmt.Errorf(errorForVersion, version)
	}

	major, err := strconv.Atoi(splitVersions[0])
	if err != nil {
		return 0, 0, fmt.Errorf(errorForVersion, version)
	}

	minor, err := strconv.Atoi(splitVersions[1])
	if err != nil {
		return 0, 0, fmt.Errorf(errorForVersion, version)
	}

	return major, minor, nil
}
