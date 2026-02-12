package download_clients

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pivotal-cf/go-pivnet/v9"
	"github.com/pivotal-cf/go-pivnet/v9/logshim"

	"github.com/pivotal-cf/go-pivnet/v9/download"
	pivnetlog "github.com/pivotal-cf/go-pivnet/v9/logger"
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

type PivnetFactory func(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) (PivnetDownloader, error)

var NewPivnetClient = func(stdout *log.Logger, stderr *log.Logger, trace bool, factory PivnetFactory, token string, skipSSL bool, pivnetHost string, proxyURL string, proxyUsername string, proxyPassword string, proxyAuthType string, proxyKrb5Config string) (ProductDownloader, error) {
	logger := logshim.NewLogShim(
		stdout,
		stderr,
		trace,
	)

	// Configure proxy settings
	proxyConfig := pivnet.ProxyAuthConfig{}
	if proxyURL != "" {
		proxyConfig.ProxyURL = proxyURL

		// Set proxy authentication if auth type is provided
		if proxyAuthType != "" {
			proxyConfig.AuthType = pivnet.ProxyAuthType(proxyAuthType)
			proxyConfig.Username = proxyUsername
			proxyConfig.Password = proxyPassword

			// Set Kerberos config file path if provided (for SPNEGO authentication)
			if proxyKrb5Config != "" {
				proxyConfig.Krb5Config = proxyKrb5Config
			}
		}
	}

	// Create token generator with proxy config
	tokenGenerator := pivnet.NewAccessTokenOrLegacyTokenWithProxy(
		token,
		pivnetHost,
		skipSSL,
		proxyConfig,
		userAgent,
	)

	// Create base config
	config := pivnet.ClientConfig{
		Host:              pivnetHost,
		UserAgent:         userAgent,
		SkipSSLValidation: skipSSL,
		ProxyAuthConfig:   proxyConfig,
	}

	// Create downloader with config (includes proxy settings if configured)
	downloader, err := factory(
		tokenGenerator,
		config,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pivnet downloader: %w", err)
	}

	// Create client with same config (includes proxy settings if configured)
	// This client is used for metadata extraction in PivnetFileArtifact
	client, err := pivnet.NewClientWithProxy(
		tokenGenerator,
		config,
		logger,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create Pivnet client: %w", err)
	}

	return &pivnetClient{
		downloader: downloader,
		stderr:     stderr,
		client:     client,
	}, nil
}

type pivnetClient struct {
	downloader PivnetDownloader
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

	productFiles, err = filterProductFilesByGlob(productFiles, []string{glob})
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

func (p *pivnetClient) GetLatestStemcellForProduct(fa FileArtifacter, _ string, stemcellSlug string) (StemcellArtifacter, error) {
	fileArtifact := fa.(*PivnetFileArtifact)
	dependencies, err := p.downloader.ReleaseDependencies(fileArtifact.slug, fileArtifact.releaseID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch stemcell dependency for %s: %s", fileArtifact.slug, err)
	}

	stemcellSlug, stemcellVersion, err := p.getLatestStemcell(dependencies, stemcellSlug)
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

func (p *pivnetClient) getLatestStemcell(dependencies []pivnet.ReleaseDependency, stemcellSlug string) (string, string, error) {
	var (
		versions            []string
		matchedStemcellSlug string
	)
	stemCellSlugVersions := make(map[string][]string)

	for _, dependency := range dependencies {
		if strings.Contains(dependency.Release.Product.Slug, "stemcells") {
			stemCellSlugVersions[dependency.Release.Product.Slug] = append(stemCellSlugVersions[dependency.Release.Product.Slug], dependency.Release.Version)
			if strings.Contains(dependency.Release.Product.Slug, stemcellSlug) {
				matchedStemcellSlug = dependency.Release.Product.Slug
			}
		}
	}

	if len(stemCellSlugVersions) >= 2 && stemcellSlug == "" {
		var stemcells []string
		for k := range stemCellSlugVersions {
			stemcells = append(stemcells, k)
		}
		return "", "", fmt.Errorf("multiple stemcell types found: %q."+
			" Provide %q option to specify the stemcell slug of stemcell that needs to be downloaded",
			strings.Join(stemcells, ","), "stemcell-slug")
	}

	if matchedStemcellSlug == "" {
		return "", "", fmt.Errorf("provided %q stemcell slug is invalid, please provide the correct one", stemcellSlug)
	}

	if stemcellSlug != "" {
		versions = stemCellSlugVersions[matchedStemcellSlug]
	} else {
		for _, stemcellVersions := range stemCellSlugVersions {
			versions = stemcellVersions
		}
	}

	stemcellVersion, err := getLatestStemcellVersion(versions)
	if err != nil {
		return "", "", err
	}
	return matchedStemcellSlug, stemcellVersion, nil
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

// filterProductFilesByGlob filters product files by matching their AWSObjectKey against a glob pattern
func filterProductFilesByGlob(productFiles []pivnet.ProductFile, globs []string) ([]pivnet.ProductFile, error) {
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

func DefaultPivnetFactory(ts pivnet.AccessTokenService, config pivnet.ClientConfig, logger pivnetlog.Logger) (PivnetDownloader, error) {
	// Always use NewClientWithProxy as it handles both proxy and non-proxy cases
	// It uses proxy authentication if AuthType is set, otherwise falls back to standard transport
	pivnetClient, err := pivnet.NewClientWithProxy(ts, config, logger)
	if err != nil {
		return nil, err
	}
	// Wrap the pivnet.Client in a gp.Client-compatible structure
	// We need to create a wrapper that implements PivnetDownloader interface
	return &pivnetDownloaderWrapper{
		client: pivnetClient,
	}, nil
}

// pivnetDownloaderWrapper wraps a pivnet.Client to implement PivnetDownloader interface
// This is needed when using NewClientWithProxy which returns pivnet.Client directly
type pivnetDownloaderWrapper struct {
	client pivnet.Client
}

func (w *pivnetDownloaderWrapper) ReleasesForProductSlug(slug string, params ...pivnet.QueryParameter) ([]pivnet.Release, error) {
	return w.client.Releases.List(slug, params...)
}

func (w *pivnetDownloaderWrapper) ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error) {
	releases, err := w.client.Releases.List(productSlug)
	if err != nil {
		return pivnet.Release{}, err
	}
	for _, r := range releases {
		if r.Version == releaseVersion {
			return w.client.Releases.Get(productSlug, r.ID)
		}
	}
	return pivnet.Release{}, fmt.Errorf("release not found for version: '%s'", releaseVersion)
}

func (w *pivnetDownloaderWrapper) ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error) {
	productFiles, err := w.client.ProductFiles.ListForRelease(productSlug, releaseID)
	if err != nil {
		return nil, err
	}
	fileGroups, err := w.client.FileGroups.ListForRelease(productSlug, releaseID)
	if err != nil {
		return nil, err
	}
	for _, fileGroup := range fileGroups {
		productFiles = append(productFiles, fileGroup.ProductFiles...)
	}
	return productFiles, nil
}

func (w *pivnetDownloaderWrapper) DownloadProductFile(location *download.FileInfo, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error {
	return w.client.ProductFiles.DownloadForRelease(location, productSlug, releaseID, productFileID, progressWriter)
}

func (w *pivnetDownloaderWrapper) ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error) {
	return w.client.ReleaseDependencies.List(productSlug, releaseID)
}

func (w *pivnetDownloaderWrapper) AcceptEULA(productSlug string, releaseID int) error {
	return w.client.EULA.Accept(productSlug, releaseID)
}
