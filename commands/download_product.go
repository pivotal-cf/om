package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/go-pivnet"
	pivnetlog "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/validator"
	"github.com/pivotal-cf/pivnet-cli/filter"
	"github.com/pivotal-cf/pivnet-cli/gp"
)

const DownloadProductOutputFilename = "download-file.json"

type outputList struct {
	ProductPath     string `json:"product_path,omitempty"`
	ProductSlug     string `json:"product_slug,omitempty"`
	StemcellPath    string `json:"stemcell_path,omitempty"`
	StemcellVersion string `json:"stemcell_version,omitempty"`
}

//go:generate counterfeiter -o ./fakes/pivnet_downloader_service.go --fake-name PivnetDownloader . PivnetDownloader
type PivnetDownloader interface {
	ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error)
	ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	DownloadProductFile(location *os.File, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
}

type PivnetFactory func(config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader

func DefaultPivnetFactory(config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader {
	return gp.NewClient(config, logger)
}

type DownloadProduct struct {
	environFunc    func() []string
	logger         pivnetlog.Logger
	progressWriter io.Writer
	pivnetFactory  PivnetFactory
	client         PivnetDownloader
	filter         *filter.Filter
	Options        struct {
		ConfigFile          string   `long:"config"                short:"c" description:"path to yml file for configuration (keys must match the following command line flags)"`
		VarsFile            []string `long:"vars-file"             short:"l" description:"Load variables from a YAML file"`
		VarsEnv             []string `long:"vars-env"                        description:"Load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)"`
		Token               string   `long:"pivnet-api-token"      short:"t" description:"API token to use when interacting with Pivnet. Can be retrieved from your profile page in Pivnet." required:"true"`
		FileGlob            string   `long:"pivnet-file-glob"      short:"f" description:"Glob to match files within Pivotal Network product to be downloaded." required:"true"`
		ProductSlug         string   `long:"pivnet-product-slug"   short:"p" description:"Path to product" required:"true"`
		ProductVersion      string   `long:"product-version"       short:"v" description:"version of the product-slug to download files from. Incompatible with --product-version-regex flag."`
		ProductVersionRegex string   `long:"product-version-regex" short:"r" description:"Regex pattern matching versions of the product-slug to download files from. Highest-versioned match will be used. Incompatible with --product-version flag."`
		OutputDir           string   `long:"output-directory"      short:"o" description:"Directory path to which the file will be outputted. File name will be preserved from Pivotal Network" required:"true"`
		Stemcell            bool     `long:"download-stemcell"               description:"No-op for backwards compatibility"`
		StemcellIaas        string   `long:"stemcell-iaas"                   description:"Download the latest available stemcell for the product for the specified iaas. for example 'vsphere' or 'vcloud' or 'openstack' or 'google' or 'azure' or 'aws'"`
	}
}

func NewDownloadProduct(environFunc func() []string, logger pivnetlog.Logger, progressWriter io.Writer, factory PivnetFactory) DownloadProduct {
	return DownloadProduct{
		environFunc:    environFunc,
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
	err := loadConfigFile(args, &c.Options, c.environFunc)
	if err != nil {
		return fmt.Errorf("could not parse download-product flags: %s", err)
	}

	if c.Options.ProductVersionRegex != "" && c.Options.ProductVersion != "" {
		return fmt.Errorf("cannot use both --product-version and --product-version-regex; please choose one or the other")
	}

	var productFileName, stemcellFileName string
	var releaseID int

	c.init()

	productVersion := c.Options.ProductVersion
	if c.Options.ProductVersionRegex != "" {
		re, err := regexp.Compile(c.Options.ProductVersionRegex)
		if err != nil {
			return fmt.Errorf("could not compile regex: %s: %s", c.Options.ProductVersionRegex, err)
		}

		releases, err := c.client.ReleasesForProductSlug(c.Options.ProductSlug)
		if err != nil {
			return err
		}

		var versions []*version.Version
		for _, release := range releases {
			if !re.MatchString(release.Version) {
				continue
			}

			v, err := version.NewVersion(release.Version)
			if err != nil {
				c.logger.Info(fmt.Sprintf("could not parse version: %s", release.Version))
				continue
			}
			versions = append(versions, v)
		}

		sort.Sort(version.Collection(versions))

		productVersion = versions[len(versions)-1].Original()
	}

	releaseID, productFileName, err = c.downloadProductFile(c.Options.ProductSlug, productVersion, c.Options.FileGlob)
	if err != nil {
		return fmt.Errorf("could not download product: %s", err)
	}

	if c.Options.StemcellIaas == "" {
		return c.writeOutputFile(productFileName, stemcellFileName, "")
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

	stemcellSlug, stemcellVersion, err := getLatestStemcell(dependencies)
	if err != nil {
		return fmt.Errorf("could not sort stemcell dependency: %s", err)
	}

	_, stemcellFileName, err = c.downloadProductFile(stemcellSlug, stemcellVersion, fmt.Sprintf("*%s*", c.Options.StemcellIaas))
	if err != nil {
		return fmt.Errorf("could not download stemcell: %s", err)
	}

	return c.writeOutputFile(productFileName, stemcellFileName, stemcellVersion)
}

func (c DownloadProduct) writeOutputFile(productFileName string, stemcellFileName string, stemcellVersion string) error {
	c.logger.Info(fmt.Sprintf("Writing a list of downloaded artifact to %s", DownloadProductOutputFilename))
	outputList := outputList{
		ProductPath:     productFileName,
		StemcellPath:    stemcellFileName,
		ProductSlug:     c.Options.ProductSlug,
		StemcellVersion: stemcellVersion,
	}

	outputFile, err := os.Create(path.Join(c.Options.OutputDir, DownloadProductOutputFilename))
	if err != nil {
		return fmt.Errorf("could not create %s: %s", DownloadProductOutputFilename, err)
	}
	defer outputFile.Close()

	return json.NewEncoder(outputFile).Encode(outputList)
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

	productFilePath := path.Join(c.Options.OutputDir, path.Base(productFileName.AWSObjectKey))

	exist, err := checkFileExists(productFilePath, productFileName.SHA256)
	if err != nil {
		return release.ID, productFilePath, err
	}

	if exist {
		c.logger.Info(fmt.Sprintf("%s already exists, skip downloading", productFilePath))
		return release.ID, productFilePath, nil
	}

	productFile, err := os.Create(productFilePath)
	if err != nil {
		return release.ID, "", fmt.Errorf("could not create file %s: %s", productFilePath, err)
	}
	defer productFile.Close()

	c.logger.Info(fmt.Sprintf("downloading %s...", productFile.Name()))
	err = c.client.DownloadProductFile(productFile, slug, release.ID, productFileName.ID, c.progressWriter)
	if err != nil {
		return release.ID, "", fmt.Errorf("could not download product file %s %s: %s", slug, version, err)
	}

	return release.ID, productFilePath, nil
}

func checkFileExists(path, expectedSum string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("failed to get file information: %s", err)
		}
	}

	validate := validator.NewSHA256Calculator()
	sum, err := validate.Checksum(path)
	if err != nil {
		return false, fmt.Errorf("failed to calculate the checksum: %s", err)
	}

	return sum == expectedSum, nil
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

func getLatestStemcell(dependencies []pivnet.ReleaseDependency) (string, string, error) {
	const errorForVersion = "versioning of stemcell dependency in unexpected format. the following version could not be parsed: %s"

	var stemcellSlug string
	var stemcellVersion string
	var stemcellVersionMajor int
	var stemcellVersionMinor int

	for _, dependency := range dependencies {
		if strings.Contains(dependency.Release.Product.Slug, "stemcells") {
			versionString := dependency.Release.Version
			splitVersions := strings.Split(versionString, ".")
			if len(splitVersions) == 1 {
				splitVersions = []string{splitVersions[0], "0"}
			}
			if len(splitVersions) != 2 {
				return stemcellSlug, stemcellVersion, fmt.Errorf(errorForVersion, versionString)
			}
			major, err := strconv.Atoi(splitVersions[0])
			if err != nil {
				return stemcellSlug, stemcellVersion, fmt.Errorf(errorForVersion, versionString)
			}
			minor, err := strconv.Atoi(splitVersions[1])
			if err != nil {
				return stemcellSlug, stemcellVersion, fmt.Errorf(errorForVersion, versionString)
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

	return stemcellSlug, stemcellVersion, nil
}
