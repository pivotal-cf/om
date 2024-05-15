package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/pivotal-cf/om/download_clients"
	"github.com/pivotal-cf/om/extractor"
	"github.com/pivotal-cf/om/validator"
)

// type PivnetOptions struct {
// 	PivnetProductSlug   string `long:"pivnet-product-slug"   short:"p"                          description:"path to product" required:"true"`
// 	PivnetDisableSSL    bool   `long:"pivnet-disable-ssl"                                       description:"whether to disable ssl validation when contacting the Pivotal Network"`
// 	PivnetToken         string `long:"pivnet-api-token"      short:"t"                          description:"API token to use when interacting with Pivnet. Can be retrieved from your profile page in Pivnet."`
// 	PivnetHost          string `long:"pivnet-host" description:"the API endpoint for Pivotal Network" default:"https://network.pivotal.io"`
// 	FileGlob            string `long:"file-glob"             short:"f"  description:"glob to match files within Pivotal Network product to be downloaded."`
// 	ProductVersion      string `long:"product-version"                                          description:"version of the product-slug to download files from. Incompatible with --product-version-regex flag."`
// 	ProductVersionRegex string `long:"product-version-regex" short:"r"                          description:"regex pattern matching versions of the product-slug to download files from. Highest-versioned match will be used. Incompatible with --product-version flag."`

// 	PivnetFileGlobSupport string `long:"pivnet-file-glob" hidden:"true"`
// }

type RMTOptions struct {
	RMTProductSlug      string `long:"rmt-product-slug"   short:"p"                          description:"path to product" required:"true"`
	RMTDisableSSL       bool   `long:"rmt-disable-ssl"                                       description:"whether to disable ssl validation when contacting the RMT"`
	RMTToken            string `long:"rmt-api-token"      short:"t"                          description:"API token to use when interacting with RMT. Can be retrieved from ???."`
	RMTHost             string `long:"rmt-host" description:"the API endpoint for Pivotal Network" default:"https://eapi-gcpstg.broadcom.com"`
	FileGlob            string `long:"file-glob"             short:"f"  description:"glob to match files within RMT product to be downloaded."`
	ProductVersion      string `long:"product-version"                                          description:"version of the product-slug to download files from. Incompatible with --product-version-regex flag."`
	ProductVersionRegex string `long:"product-version-regex" short:"r"                          description:"regex pattern matching versions of the product-slug to download files from. Highest-versioned match will be used. Incompatible with --product-version flag."`

	RMTFileGlobSupport string `long:"rmt-file-glob" hidden:"true"`
}

type GCSOptions struct {
	GCSServiceAccountJSON string `long:"gcs-service-account-json" description:"the service account key JSON"`
	GCSProjectID          string `long:"gcs-project-id"           description:"the project id for the bucket's gcp account"`

	GCPServiceAccountSupport string `long:"gcp-service-account-json" hidden:"true"`
	GCPProjectIDSupport      string `long:"gcp-project-id"           hidden:"true"`
}

type S3Options struct {
	S3AccessKeyID     string `long:"s3-access-key-id"                 description:"access key for the s3 compatible blobstore"`
	S3AuthType        string `long:"s3-auth-type"                     description:"can be set to \"iam\" in order to allow use of instance credentials" default:"accesskey"`
	S3SecretAccessKey string `long:"s3-secret-access-key"             description:"secret key for the s3 compatible blobstore"`
	S3RegionName      string `long:"s3-region-name"                   description:"bucket region in the s3 compatible blobstore. If not using AWS, this value is 'region'"`
	S3Endpoint        string `long:"s3-endpoint"                      description:"the endpoint to access the s3 compatible blobstore. If not using AWS, this is required"`
	S3DisableSSL      bool   `long:"s3-disable-ssl"                   description:"whether to disable ssl (https or http) when contacting the s3 compatible blobstore"`
	S3EnableV2Signing bool   `long:"s3-enable-v2-signing"             description:"whether to use v2 signing with your s3 compatible blobstore. (if you don't know what this is, leave blank, or set to 'false')"`
}

type AzureOptions struct {
	AzureStorageAccount string `long:"azure-storage-account" description:"the name of the storage account where the container exists"`
	AzureKey            string `long:"azure-storage-key"     description:"the access key for the storage account"`
}

type StemcellOptions struct {
	StemcellIaas    string `long:"stemcell-iaas"     description:"download the latest available stemcell for the product for the specified iaas. for example 'vsphere' or 'vcloud' or 'openstack' or 'google' or 'azure' or 'aws'. Can contain globbing patterns to match specific files in a stemcell release on RMT"`
	StemcellVersion string `long:"stemcell-version" description:"the version number of the stemcell to download (ie 458.61)"`
	StemcellHeavy   bool   `long:"stemcell-heavy" description:"force the downloading of a heavy stemcell, will fail if non exists"`
}

type DownloadProductOptions struct {
	Source            string `long:"source"                     short:"s" description:"enables download from external sources when set to [s3|gcs|azure|rmt]" default:"rmt"`
	OutputDir         string `long:"output-directory"           short:"o" description:"directory path to which the file will be outputted. File Name will be preserved from Pivotal Network" required:"true"`
	StemcellOutputDir string `long:"stemcell-output-directory" short:"d" description:"directory path to which the stemcell file will be outputted. If not provided, output-directory will be used."`

	Bucket               string `long:"blobstore-bucket" description:"bucket name where the product resides in the s3|gcs|azure compatible blobstore"`
	ProductPath          string `long:"blobstore-product-path"   description:"specify the lookup path where the s3|gcs|azure product artifacts are stored"`
	StemcellPath         string `long:"blobstore-stemcell-path" description:"specify the lookup path where the s3|gcs|azure stemcell artifacts are stored"`
	CacheCleanup         string `long:"cache-cleanup" env:"CACHE_CLEANUP" description:"Delete everything except the latest artifact in output-dir and stemcell-output-dir, set to 'I acknowledge this will delete files in the output directories' to accept these terms"`
	CheckAlreadyUploaded bool   `long:"check-already-uploaded" description:"Check if product is already uploaded on Ops Manager before downloading. This command is authenticated."`

	S3BucketSupport          string `long:"s3-bucket" hidden:"true"`
	GCSBucketSupport         string `long:"gcs-bucket" hidden:"true"`
	AzureBucketSupport       string `long:"azure-container" hidden:"true"`
	S3ProductPathSupport     string `long:"s3-product-path" hidden:"true"`
	GCSProductPathSupport    string `long:"gcs-product-path" hidden:"true"`
	AzureProductPathSupport  string `long:"azure-product-path" hidden:"true"`
	S3StemcellPathSupport    string `long:"s3-stemcell-path" hidden:"true"`
	GCSStemcellPathSupport   string `long:"gcs-stemcell-path" hidden:"true"`
	AzureStemcellPathSupport string `long:"azure-stemcell-path" hidden:"true"`

	AzureOptions
	GCSOptions
	InterpolateOptions interpolateConfigFileOptions `group:"config file interpolation"`
	// PivnetOptions
	RMTOptions
	S3Options
	StemcellOptions
}

type DownloadProduct struct {
	environFunc    func() []string
	progressWriter io.Writer
	stderr         *log.Logger
	stdout         *log.Logger
	service        downloadProductService
	downloadClient download_clients.ProductDownloader
	Options        DownloadProductOptions
}

//counterfeiter:generate -o ./fakes/download_product_service.go --fake-name DownloadProductService . downloadProductService
type downloadProductService interface {
	CheckProductAvailability(string, string) (bool, error)
	CheckStemcellAvailability(string) (bool, error)
}

func NewDownloadProduct(environFunc func() []string, stdout *log.Logger, stderr *log.Logger, progressWriter io.Writer, downloadProductService downloadProductService) *DownloadProduct {
	return &DownloadProduct{
		environFunc:    environFunc,
		stderr:         stderr,
		stdout:         stdout,
		progressWriter: progressWriter,
		service:        downloadProductService,
	}
}

func (c *DownloadProduct) Execute(args []string) error {
	err := c.validate()
	fmt.Printf("\nValidated\n")
	if err != nil {
		return err
	}

	err = c.createClient()
	fmt.Printf("\nClient Created\n")
	if err != nil {
		return err
	}

	productVersion, err := c.determineProductVersion()
	fmt.Printf("\nProduct Version Determined\n")
	if err != nil {
		return err
	}

	productFileName, productFileArtifact, err := c.downloadProductFile(
		c.Options.RMTProductSlug,
		productVersion,
		c.Options.FileGlob,
		fmt.Sprintf("[%s,%s]", c.Options.RMTProductSlug, productVersion),
		c.Options.OutputDir,
	)

	if err != nil {
		return fmt.Errorf("could not download product: %s", err)
	}
	fmt.Printf("\nProduct Downloaded\n")
	if c.Options.StemcellIaas == "" {
		return c.writeDownloadProductOutput(productFileName, productVersion, "", "")
	}

	if filepath.Ext(productFileName) != ".pivotal" {
		c.stderr.Printf("the downloaded file is not a .pivotal file. Not determining and fetching required stemcell.")
		return c.writeDownloadProductOutput(productFileName, productVersion, "", "")
	}

	stemcellVersion, stemcellFileName, err := c.downloadStemcell(productFileName, productVersion, productFileArtifact)
	if err != nil {
		return err
	}

	err = c.writeDownloadProductOutput(productFileName, productVersion, stemcellFileName, stemcellVersion)
	if err != nil {
		return err
	}

	return c.writeAssignStemcellInput(productFileName, productFileArtifact, stemcellVersion)
}

func (c *DownloadProduct) downloadStemcell(productFileName string, productVersion string, productFileArtifact download_clients.FileArtifacter) (string, string, error) {
	c.stderr.Printf("Downloading stemcell")

	stemcell, err := c.downloadClient.GetLatestStemcellForProduct(productFileArtifact, productFileName)
	if err != nil {
		return "", "", fmt.Errorf("could not get information about stemcell: %s", err)
	}

	stemcellGlobs := []string{
		fmt.Sprintf("light*bosh*%s*", c.Options.StemcellIaas),
		fmt.Sprintf("bosh*%s*", c.Options.StemcellIaas),
	}

	if c.Options.StemcellHeavy {
		stemcellGlobs = []string{
			fmt.Sprintf("bosh*%s*", c.Options.StemcellIaas),
		}
	}

	stemcellVersion := stemcell.Version()
	if c.Options.StemcellVersion != "" {
		stemcellVersion = c.Options.StemcellVersion
	}

	stemcellFileName, err := "", nil

	var stemcellOutputDirectory string

	if c.Options.StemcellOutputDir != "" {
		stemcellOutputDirectory = c.Options.StemcellOutputDir
	} else {
		stemcellOutputDirectory = c.Options.OutputDir
	}

	for _, stemcellGlob := range stemcellGlobs {
		stemcellFileName, _, err = c.downloadProductFile(
			stemcell.Slug(),
			stemcellVersion,
			stemcellGlob,
			fmt.Sprintf("[%s,%s]", stemcell.Slug(), stemcellVersion),
			stemcellOutputDirectory,
		)
		if err == nil {
			break
		}
	}

	if err != nil {
		isHeavy := ""
		if c.Options.StemcellHeavy {
			isHeavy = "heavy "
		}
		return "", "", fmt.Errorf("could not download stemcell: %s\nNo %sstemcell identified for IaaS \"%s\" on Pivotal Network. Correct the `stemcell-iaas` option to match the IaaS portion of the stemcell filename, or remove the option.", err, isHeavy, c.Options.StemcellIaas)
	}
	return stemcellVersion, stemcellFileName, nil
}

func (c *DownloadProduct) determineProductVersion() (string, error) {
	return download_clients.DetermineProductVersion(
		c.Options.RMTProductSlug,
		c.Options.ProductVersion,
		c.Options.ProductVersionRegex,
		c.downloadClient,
		c.stderr,
	)
}

func (c *DownloadProduct) createClient() error {
	plugin, err := newDownloadClientFromSource(c.Options, c.progressWriter, c.stdout, c.stderr)
	if err != nil {
		return fmt.Errorf("could not find valid source for '%s': %w", c.Options.Source, err)
	}

	c.downloadClient = plugin
	return nil
}

func (c *DownloadProduct) validate() error {
	c.handleAliases()

	if c.Options.FileGlob == "" {
		return errors.New("--file-glob is required")
	}

	if c.Options.ProductVersionRegex != "" && c.Options.ProductVersion != "" {
		return errors.New("cannot use both --product-version and --product-version-regex; please choose one or the other")
	}

	if c.Options.ProductVersionRegex == "" && c.Options.ProductVersion == "" {
		return errors.New("no version information provided; please provide either --product-version or --product-version-regex")
	}

	// if c.Options.PivnetToken == "" && c.Options.Source == "pivnet" {
	// 	return errors.New(`could not execute "download-product": could not parse download-product flags: missing required flag "--pivnet-api-token"`)
	// }
	if c.Options.RMTToken == "" && (c.Options.Source == "rmt" || c.Options.Source == "RMT") {
		return errors.New(`could not execute "download-product": could not parse download-product flags: missing required flag "--rmt-api-token"`)
	}

	if c.Options.StemcellHeavy && c.Options.StemcellIaas == "" {
		return errors.New("--stemcell-heavy requires --stemcell-iaas to be defined")
	}
	if c.Options.StemcellVersion != "" && c.Options.StemcellIaas == "" {
		return errors.New("--stemcell-version requires --stemcell-iaas to be defined")
	}

	file, err := os.Open(c.Options.OutputDir)
	if err != nil {
		return fmt.Errorf("--output-directory %q does not exist: %w", c.Options.OutputDir, err)
	}
	fi, _ := file.Stat()
	if !fi.IsDir() {
		return fmt.Errorf("--output-directory %q is not a directory", c.Options.OutputDir)
	}

	if c.Options.StemcellOutputDir != "" {
		file, err = os.Open(c.Options.StemcellOutputDir)
		if err != nil {
			return fmt.Errorf("--stemcell-output-directory %q does not exist: %w", c.Options.StemcellOutputDir, err)
		}
		fi, _ = file.Stat()
		if !fi.IsDir() {
			return fmt.Errorf("--stemcell-output-directory %q is not a directory", c.Options.StemcellOutputDir)
		}
	}
	return nil
}

func (c *DownloadProduct) handleAliases() {
	if c.Options.S3BucketSupport != "" {
		c.Options.Bucket = c.Options.S3BucketSupport
	}
	if c.Options.GCSBucketSupport != "" {
		c.Options.Bucket = c.Options.GCSBucketSupport
	}
	if c.Options.AzureBucketSupport != "" {
		c.Options.Bucket = c.Options.AzureBucketSupport
	}

	if c.Options.S3ProductPathSupport != "" {
		c.Options.ProductPath = c.Options.S3ProductPathSupport
	}
	if c.Options.GCSProductPathSupport != "" {
		c.Options.ProductPath = c.Options.GCSProductPathSupport
	}
	if c.Options.AzureProductPathSupport != "" {
		c.Options.ProductPath = c.Options.AzureProductPathSupport
	}

	if c.Options.S3StemcellPathSupport != "" {
		c.Options.StemcellPath = c.Options.S3StemcellPathSupport
	}
	if c.Options.GCSStemcellPathSupport != "" {
		c.Options.StemcellPath = c.Options.GCSStemcellPathSupport
	}
	if c.Options.AzureStemcellPathSupport != "" {
		c.Options.StemcellPath = c.Options.AzureStemcellPathSupport
	}
	if c.Options.RMTFileGlobSupport != "" {
		c.Options.FileGlob = c.Options.RMTFileGlobSupport
	}
	if c.Options.GCPServiceAccountSupport != "" {
		c.Options.GCSServiceAccountJSON = c.Options.GCPServiceAccountSupport
	}
	if c.Options.GCPProjectIDSupport != "" {
		c.Options.GCSProjectID = c.Options.GCPProjectIDSupport
	}
}

func (c DownloadProduct) writeDownloadProductOutput(productFileName string, productVersion string, stemcellFileName string, stemcellVersion string) error {
	downloadProductFilename := "download-file.json"
	c.stderr.Printf("Writing a list of downloaded artifact to %s", downloadProductFilename)
	downloadProductPayload := struct {
		ProductPath     string `json:"product_path,omitempty"`
		ProductSlug     string `json:"product_slug,omitempty"`
		ProductVersion  string `json:"product_version,omitempty"`
		StemcellPath    string `json:"stemcell_path,omitempty"`
		StemcellVersion string `json:"stemcell_version,omitempty"`
	}{
		ProductPath:     productFileName,
		StemcellPath:    stemcellFileName,
		ProductSlug:     c.Options.RMTProductSlug,
		ProductVersion:  productVersion,
		StemcellVersion: stemcellVersion,
	}

	outputFile, err := os.Create(filepath.Join(c.Options.OutputDir, downloadProductFilename))
	if err != nil {
		return fmt.Errorf("could not create %s: %s", downloadProductFilename, err)
	}
	defer outputFile.Close()

	err = json.NewEncoder(outputFile).Encode(downloadProductPayload)
	if err != nil {
		return fmt.Errorf("could not encode JSON for %s: %s", downloadProductFilename, err)
	}

	return nil
}

func (c DownloadProduct) writeAssignStemcellInput(productFile string, fileArtifact download_clients.FileArtifacter, stemcellVersion string) error {
	assignStemcellFileName := "assign-stemcell.yml"

	var (
		metadata *extractor.Metadata
		err      error
	)

	if c.Options.CheckAlreadyUploaded {
		metadata, err = fileArtifact.ProductMetadata()
		if err != nil {
			if !errors.Is(err, download_clients.ErrCannotExtractMetadata) {
				return fmt.Errorf("cannot parse product metadata: %s", err)
			}

			c.stderr.Printf("cannot extract metadata because the product file was not downloaded, will not create assign-stemcell input: %v", err)
			return nil
		}
	} else {
		metadataExtractor := extractor.NewMetadataExtractor()
		if metadata, err = metadataExtractor.ExtractFromFile(productFile); err != nil {
			return fmt.Errorf("cannot parse product metadata: %s", err)
		}
	}

	c.stderr.Printf("Writing a assign stemcell artifact to %s", assignStemcellFileName)

	assignStemcellPayload := struct {
		Product  string `json:"product"`
		Stemcell string `json:"stemcell"`
	}{
		Product:  metadata.Name,
		Stemcell: stemcellVersion,
	}

	outputFile, err := os.Create(filepath.Join(c.Options.OutputDir, assignStemcellFileName))
	if err != nil {
		return fmt.Errorf("could not create %s: %s", assignStemcellFileName, err)
	}
	defer outputFile.Close()

	err = json.NewEncoder(outputFile).Encode(assignStemcellPayload)
	if err != nil {
		return fmt.Errorf("could not encode JSON for %s: %s", assignStemcellFileName, err)
	}

	return nil
}

func (c *DownloadProduct) downloadProductFile(slug, version, glob, prefixPath string, outputDir string) (string, download_clients.FileArtifacter, error) {
	fileArtifact, err := c.downloadClient.GetLatestProductFile(slug, version, glob)

	if err != nil {
		fmt.Printf("\nFailed line 418\n")
		return "", nil, err
	}

	var productFilePath string
	if c.Options.Source != "rmt" || c.Options.Bucket == "" {
		productFilePath = filepath.Join(outputDir, filepath.Base(fileArtifact.Name()))
	} else {
		productFilePath = filepath.Join(outputDir, prefixPath+filepath.Base(fileArtifact.Name()))
	}

	c.stderr.Printf("attempting to download the file %s from source %s", fileArtifact.Name(), c.downloadClient.Name())

	// check for already downloaded file
	exist, err := checkFileExists(productFilePath)
	if err != nil {
		fmt.Printf("\nFailed line 434\n")
		return productFilePath, nil, err
	}

	if c.Options.CheckAlreadyUploaded {
		if filepath.Ext(fileArtifact.Name()) == ".pivotal" {
			c.stderr.Printf("checking if product already available on Ops Manager...")

			metadata, err := fileArtifact.ProductMetadata()
			if err != nil {
				fmt.Printf("\nFailed line 444\n")
				return "", nil, fmt.Errorf("could not extract metadata from product: %w", err)
			}

			found, err := c.service.CheckProductAvailability(metadata.Name, metadata.Version)
			if err != nil {
				c.stderr.Printf("could not determine if product is already uploaded")
				fmt.Printf("\nFailed line 451\n")
				return "", nil, fmt.Errorf("could not check Ops Manager for product: %w", err)
			}

			if found {
				c.stderr.Println("product found. Skipping download.")
				fmt.Printf("\nFailed Line 457\n")
				return productFilePath, fileArtifact, nil
			}
			c.stderr.Println("product not found. Continuing download...")
		} else {
			c.stderr.Printf("checking if stemcell already available on Ops Manager...")

			filename := filepath.Base(fileArtifact.Name())
			found, err := c.service.CheckStemcellAvailability(filename)
			if err != nil {
				c.stderr.Printf("could not determine if stemcell is already uploaded")
				fmt.Printf("\nfailed 468\n")
				return "", nil, fmt.Errorf("could not check Ops Manager for stemcell: %w", err)
			}

			if found {
				c.stderr.Println("stemcell found. Skipping download.")
				fmt.Printf("\nfailed 474\n")
				return productFilePath, fileArtifact, nil
			}
			c.stderr.Println("stemcell not found. Continuing download...")
		}
	}

	if exist {
		c.stderr.Printf("%s already exists, skip downloading", productFilePath)

		err = c.cleanupCacheArtifacts(outputDir, glob, productFilePath, slug)
		if err != nil {
			fmt.Printf("\nfailed 486\n")
			return "", nil, fmt.Errorf("could not cleanup cache: %w", err)
		}

		return productFilePath, fileArtifact, nil
	}

	err = c.cleanupCacheArtifacts(outputDir, glob, productFilePath, slug)
	if err != nil {
		fmt.Printf("\nfailed 495\n")
		return "", nil, fmt.Errorf("could not cleanup cache: %w", err)
	}

	partialProductFilePath := productFilePath + ".partial"
	// create a new file to download
	productFile, err := os.Create(partialProductFilePath)
	if err != nil {
		fmt.Printf("\nfailed 503\n")
		return "", nil, fmt.Errorf("could not create file %s: %s", productFilePath, err)
	}
	defer productFile.Close()

	err = c.downloadClient.DownloadProductToFile(fileArtifact, productFile)
	if err != nil {
		fmt.Printf("\nFailed 510\n")
		return productFilePath, fileArtifact, err
	}

	// check for correct sha on newly downloaded file
	if ok, calculatedSum := c.shasumMatches(partialProductFilePath, fileArtifact.SHA256()); !ok {
		e := fmt.Sprintf("the sha (%s) from %s does not match the calculated sha (%s) for the file %s",
			fileArtifact.SHA256(),
			c.downloadClient.Name(),
			calculatedSum,
			productFilePath)
		c.stderr.Print(e)
		_ = os.Remove(partialProductFilePath)
		fmt.Printf("\nfailed\n")
		return productFilePath, fileArtifact, errors.New(e)
	}

	_ = os.Rename(partialProductFilePath, productFilePath)
	return productFilePath, fileArtifact, nil
}

func (c *DownloadProduct) cleanupCacheArtifacts(outputDir string, glob string, productFilePath string, slug string) error {
	if c.Options.CacheCleanup == "I acknowledge this will delete files in the output directories" {

		outputDirContents, err := ioutil.ReadDir(outputDir)
		if err != nil {
			return err
		}

		prefixedGlob := fmt.Sprintf("\\[%s,*\\]%s", slug, glob)
		globs := []string{glob, prefixedGlob}
		for _, fileGlob := range globs {
			c.stderr.Printf("Cleaning up cached artifacts in directory '%s' with the glob '%s'", outputDir, fileGlob)
			for _, file := range outputDirContents {
				dirFilePath := path.Join(outputDir, file.Name())
				c.stderr.Printf("checking if %q needs to cleaned up", file.Name())
				if matchGlob, _ := filepath.Match(fileGlob, file.Name()); matchGlob {
					if dirFilePath != productFilePath {
						c.stderr.Printf("cleaning up cached file: %s", dirFilePath)
						_ = os.Remove(dirFilePath)
					}
				}
			}
		}
	}

	return nil
}

func (c *DownloadProduct) shasumMatches(path, exepectedSum string) (bool, string) {
	if exepectedSum == "" {
		return true, ""
	}

	c.stderr.Printf("calculating sha sum for %s", path)
	validate := validator.NewSHA256Calculator()
	calculatedSum, err := validate.Checksum(path)
	if err != nil {
		return false, ""
	}

	return calculatedSum == exepectedSum, calculatedSum
}

func checkFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("failed to get file information: %s", err)
		}
	}

	return true, nil
}

type ProductClientRegistration func(
	c DownloadProductOptions,
	progressWriter io.Writer,
	stdout *log.Logger,
	stderr *log.Logger,
) (download_clients.ProductDownloader, error)

func newDownloadClientFromSource(c DownloadProductOptions,
	progressWriter io.Writer,
	stdout *log.Logger,
	stderr *log.Logger,
) (download_clients.ProductDownloader, error) {
	switch c.Source {
	case "azure":
		return download_clients.NewAzureClient(
			download_clients.StowWrapper{},
			download_clients.AzureConfiguration{
				Container:      c.Bucket,
				StorageAccount: c.AzureStorageAccount,
				Key:            c.AzureKey,
				ProductPath:    c.ProductPath,
				StemcellPath:   c.StemcellPath,
			},
			stderr,
		)
	case "gcs":
		return download_clients.NewGCSClient(
			download_clients.StowWrapper{},
			download_clients.GCSConfiguration{
				Bucket:             c.Bucket,
				ProjectID:          c.GCSProjectID,
				ServiceAccountJSON: c.GCSServiceAccountJSON,
				ProductPath:        c.ProductPath,
				StemcellPath:       c.StemcellPath,
			},
			stderr,
		)
	case "s3":
		return download_clients.NewS3Client(
			download_clients.StowWrapper{},
			download_clients.S3Configuration{
				Bucket:          c.Bucket,
				AccessKeyID:     c.S3AccessKeyID,
				AuthType:        c.S3AuthType,
				SecretAccessKey: c.S3SecretAccessKey,
				RegionName:      c.S3RegionName,
				Endpoint:        c.S3Endpoint,
				DisableSSL:      c.S3DisableSSL,
				EnableV2Signing: c.S3EnableV2Signing,
				ProductPath:     c.ProductPath,
				StemcellPath:    c.StemcellPath,
			},
			stderr,
		)
	case "rmt", "RMT", "":
		return download_clients.NewPivnetClient(
			stdout,
			stderr,
			download_clients.DefaultPivnetFactory,
			c.RMTToken,
			c.RMTDisableSSL,
			c.RMTHost,
		), nil
	}

	return nil, errors.New("could not find a plugin")
}
