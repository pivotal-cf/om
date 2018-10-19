package commands

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet"
	log "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/pivnet-cli/filter"
	"github.com/pivotal-cf/pivnet-cli/gp"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

//go:generate counterfeiter -o ./fakes/pivnet_downloader_service.go --fake-name PivnetDownloader . PivnetDownloader
type PivnetDownloader interface {
	ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	DownloadProductFile(location *os.File, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
}

type PivnetFactory func(config pivnet.ClientConfig, logger log.Logger) PivnetDownloader

func DefaultPivnetFactory(config pivnet.ClientConfig, logger log.Logger) PivnetDownloader {
	return gp.NewClient(config, logger)
}

type DownloadProduct struct {
	logger         log.Logger
	progressWriter io.Writer
	pivnetFactory  PivnetFactory
	client         PivnetDownloader
	filter         *filter.Filter
	Options        struct {
		ConfigFile     string `long:"config"               short:"c"   description:"path to yml file for configuration (keys must match the following command line flags)"`
		Token          string `long:"pivnet-api-token"                  required:"true"`
		FileGlob       string `long:"pivnet-file-glob"     short:"f"   description:"Glob to match files within Pivotal Network product to be downloaded." required:"true"`
		ProductSlug    string `long:"pivnet-product-slug"  short:"p"   description:"Path to product" required:"true"`
		ProductVersion string `long:"product-version"                  description:"version of the provided product file to be used for validation" required:"true"`
		OutputDir      string `long:"output-directory"     short:"o"   description:"Directory path to which the file will be outputted. File name will be preserved from Pivotal Network" required:"true"`
		Stemcell       bool   `long:"download-stemcell"                description:"If set, the latest available stemcell for the product will also be downloaded"`
		StemcellIaas   string `long:"stemcell-iaas"                    description:"The stemcell for the specified iaas. for example 'vsphere' or 'vcloud' or 'openstack' or 'google' or 'azure' or 'aws'"`
	}
}

func NewDownloadProduct(logger log.Logger, progressWriter io.Writer, factory PivnetFactory) DownloadProduct {
	return DownloadProduct{
		logger:         logger,
		progressWriter: progressWriter,
		pivnetFactory:  factory,
		filter:         filter.NewFilter(logger),
	}
}

func (c DownloadProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
		ShortDescription: "downloads a specified product file from Pivotal Network",
		Flags:            c.Options,
	}
}

func (c DownloadProduct) Execute(args []string) error {
	err := loadConfigFile(args, &c.Options)
	if err != nil {
		return fmt.Errorf("could not parse download-product flags: %s", err)
	}

	c.init()

	releaseID, productFileName, err := c.downloadProductFile(c.Options.ProductSlug, c.Options.ProductVersion, c.Options.FileGlob)
	if err != nil {
		return fmt.Errorf("could not download product: %s", err)
	}

	if !c.Options.Stemcell {
		return nil
	}

	c.logger.Info("Downloading stemcell")

	nameParts := strings.Split(productFileName, ".")
	if nameParts[len(nameParts)-1] != "pivotal" {
		c.logger.Info("the downloaded file is not a .pivotal file. Not determining and fetching required stemcell.")
		return nil
	}

	dependencies, err := c.client.ReleaseDependencies(c.Options.ProductSlug, releaseID)
	if err != nil {
		return fmt.Errorf("could not fetch stemcell dependency for %s %s: %s", c.Options.ProductSlug, c.Options.ProductVersion, err)
	}

	var stemcellSlug string
	var stemcellVersion string
	var stemcellVersionMajor int
	var stemcellVersionMinor int

	const errorForVersion = "versioning of stemcell dependency in unexpected format. Please contact Pivotal Support and advise them that the following version could not be parsed: %s"

	for _, dependency := range dependencies {
		if strings.Contains(dependency.Release.Product.Slug, "stemcells") {
			versionString := dependency.Release.Version
			splitVersions := strings.Split(versionString, ".")
			if len(splitVersions) != 2 {
				return fmt.Errorf(errorForVersion, versionString)
			}
			major, err := strconv.Atoi(splitVersions[0])
			if err != nil {
				return fmt.Errorf(errorForVersion, versionString)
			}
			minor, err := strconv.Atoi(splitVersions[1])
			if err != nil {
				return fmt.Errorf(errorForVersion, versionString)
			}
			if major > stemcellVersionMajor {
				stemcellVersionMajor = major
				stemcellVersionMinor = minor
				stemcellVersion = versionString
				stemcellSlug = dependency.Release.Product.Slug
			} else if major == stemcellVersionMajor && minor > stemcellVersionMinor {
				stemcellVersionMinor = minor
				stemcellVersion = versionString
				stemcellSlug = dependency.Release.Product.Slug
			}
		}
	}

	_, _, err = c.downloadProductFile(stemcellSlug, stemcellVersion, fmt.Sprintf("*%s*", c.Options.StemcellIaas))
	if err != nil {
		return fmt.Errorf("could not download stemcell: %s", err)
	}

	return nil
}

func (c *DownloadProduct) init() {
	c.client = c.pivnetFactory(
		pivnet.ClientConfig{
			Host:              pivnet.DefaultHost,
			Token:             c.Options.Token,
			UserAgent:         fmt.Sprintf("om-download-product"),
			SkipSSLValidation: false,
		},
		c.logger,
	)
}

func (c *DownloadProduct) downloadProductFile(slug, version, glob string) (int, string, error) {
	release, err := c.client.ReleaseForVersion(slug, version)
	if err != nil {
		return release.ID, "", fmt.Errorf("could not fetch the release for %s %s: %s", slug, version, err)
	}

	productFileNames, err := c.client.ProductFilesForRelease(slug, release.ID)
	if err != nil {
		return release.ID, "", fmt.Errorf("could not fetch the product files for %s %s: %s", slug, version, err)
	}

	productFileNames, err = c.filter.ProductFileKeysByGlobs(productFileNames, []string{glob})
	if err != nil {
		return release.ID, "", fmt.Errorf("could not glob product files: %s", err)
	}

	if err := checkSingleProductFile(glob, productFileNames); err != nil {
		return release.ID, "", err
	}

	productFileName := productFileNames[0]

	productFile, err := os.Create(path.Join(c.Options.OutputDir, path.Base(productFileName.AWSObjectKey)))
	if err != nil {
		return release.ID, "", fmt.Errorf("could not create file %s: %s", productFile.Name(), err)
	}
	defer productFile.Close()

	err = c.client.DownloadProductFile(productFile, slug, release.ID, productFileName.ID, c.progressWriter)
	if err != nil {
		return release.ID, "", fmt.Errorf("could not download product file %s %s: %s", slug, version, err)
	}

	return release.ID, productFile.Name(), nil
}

func checkSingleProductFile(glob string, productFiles []pivnet.ProductFile) error {
	if len(productFiles) > 1 {
		var productFileNames []string
		for _, productFile := range productFiles {
			productFileNames = append(productFileNames, path.Base(productFile.AWSObjectKey))
		}
		return fmt.Errorf("the glob '%s' matches multiple files. Write your glob to match exactly one of the following:\n  %s", glob, strings.Join(productFileNames, "\n  "))
	} else if len(productFiles) == 0 {
		return fmt.Errorf("the glob '%s' matchs no file", glob)
	}

	return nil
}
