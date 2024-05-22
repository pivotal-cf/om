package download_clients

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pivotal-cf/go-pivnet/v6/logshim"
	"github.com/pivotal-cf/pivnet-cli/v2/filter"

	"github.com/pivotal-cf/go-pivnet/v6"
	"github.com/pivotal-cf/go-pivnet/v6/download"
	pivnetlog "github.com/pivotal-cf/go-pivnet/v6/logger"
	"github.com/pivotal-cf/pivnet-cli/v2/gp"
)

//counterfeiter:generate -o ./fakes/pivnet_downloader_service.go --fake-name PivnetDownloader . PivnetDownloader
type PivnetDownloader interface {
	ReleasesForProductSlug(string, ...pivnet.QueryParameter) ([]pivnet.Release, error)
	ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	DownloadProductFile(location *download.FileInfo, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
	AcceptEULA(productSlug string, releaseID int) error
}

type PivnetFactory func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader

var NewPivnetClient = func(stdout *log.Logger, stderr *log.Logger, factory PivnetFactory, token string, skipSSL bool, pivnetHost string) ProductDownloader {
	logger := logshim.NewLogShim(
		stdout,
		stderr,
		false,
	)
	tokenGenerator := pivnet.NewAccessTokenOrLegacyToken(
		token,
		pivnetHost,
		skipSSL,
		userAgent,
	)
	config := pivnet.ClientConfig{
		Host:              pivnetHost,
		UserAgent:         userAgent,
		SkipSSLValidation: skipSSL,
	}
	downloader := factory(
		tokenGenerator,
		config,
		logger,
	)
	client := pivnet.NewClient(
		tokenGenerator,
		config,
		logger,
	)

	return &pivnetClient{
		filter:     filter.NewFilter(logger),
		downloader: downloader,
		stderr:     stderr,
		client:     client,
	}
}

type pivnetClient struct {
	downloader PivnetDownloader
	filter     *filter.Filter
	stderr     *log.Logger
	client     pivnet.Client
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

func (p *pivnetClient) Name() string {
	return "pivnet"
}

func (p *pivnetClient) GetLatestProductFile(slug, version, glob string) (FileArtifacter, error) {
	// 1. Check the release for given version / slug
	release, err := p.downloader.ReleaseForVersion(slug, version)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the release for %s %s: %s", slug, version, err)
	}

	err = p.downloader.AcceptEULA(slug, release.ID)
	if err != nil {
		return nil, fmt.Errorf("could not accept EULA for download product file %s at version %s: %s", slug, version, err)
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
		return nil, fmt.Errorf("for product version %s: %s", version, err)
	}

	return &PivnetFileArtifact{
		releaseID:   release.ID,
		slug:        slug,
		productFile: productFiles[0],
		client:      p.client,
	}, nil
}

func (p *pivnetClient) DownloadProductToFile(fa FileArtifacter, file *os.File) error {
	fileArtifact := fa.(*PivnetFileArtifact)
	fileInfo, err := download.NewFileInfo(file)
	if err != nil {
		return fmt.Errorf("could not create fileInfo for download product file %s: %s", fileArtifact.slug, err.Error())
	}
	err = p.downloader.DownloadProductFile(fileInfo, fileArtifact.slug, fileArtifact.releaseID, fileArtifact.productFile.ID, p.stderr.Writer())
	if err != nil {
		return fmt.Errorf("could not download product file %s: %s", fileArtifact.slug, err)
	}
	return nil
}

func (p *pivnetClient) GetLatestStemcellForProduct(fa FileArtifacter, _ string) (StemcellArtifacter, error) {
	fileArtifact := fa.(*PivnetFileArtifact)
	dependencies, err := p.downloader.ReleaseDependencies(fileArtifact.slug, fileArtifact.releaseID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch stemcell dependency for %s: %s", fileArtifact.slug, err)
	}

	stemcellSlug, stemcellVersion, err := p.getLatestStemcell(dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not sort stemcell dependency: %s", err)
	}

	return &stemcell{
		slug:    stemcellSlug,
		version: stemcellVersion,
	}, nil
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

const (
	errorForVersion = "versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: %s"
	userAgent       = "om-download-product"
)

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

func DefaultPivnetFactory(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader {
	return gp.NewClient(ts, config, logger)
}
