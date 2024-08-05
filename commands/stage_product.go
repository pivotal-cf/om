package commands

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
)

type StageProduct struct {
	logger  logger
	service stageProductService
	Options struct {
		ConfigFile string `long:"config" description:"the config file to load product name and version (can be same as the product configuration file)"`
		Product    string `yaml:"product-name" long:"product-name"    short:"p" description:"name of product"`
		Version    string `yaml:"product-version" long:"product-version"        description:"version of product"`
	}
}

//counterfeiter:generate -o ./fakes/stage_product_service.go --fake-name StageProductService . stageProductService
type stageProductService interface {
	CheckProductAvailability(productName string, productVersion string) (bool, error)
	GetDiagnosticReport() (api.DiagnosticReport, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	Stage(api.StageProductInput, string) error
	GetLatestAvailableVersion(productName string) (string, error)
}

func NewStageProduct(service stageProductService, logger logger) *StageProduct {
	return &StageProduct{
		logger:  logger,
		service: service,
	}
}

func (sp StageProduct) Execute(args []string) error {
	err := sp.loadConfig()
	if err != nil {
		return err
	}

	productName := sp.Options.Product
	productVersion := sp.Options.Version

	err = checkRunningInstallation(sp.service.ListInstallations)
	if err != nil {
		return err
	}

	diagnosticReport, err := sp.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	deployedProductGUID := ""
	deployedProducts, err := sp.service.ListDeployedProducts()

	for _, deployedProduct := range deployedProducts {
		if deployedProduct.Type == productName {
			deployedProductGUID = deployedProduct.GUID
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	if productVersion == "latest" {
		latestVersion, err := sp.service.GetLatestAvailableVersion(productName)
		if err != nil {
			return fmt.Errorf("could not find latest version: %w", err)
		}
		productVersion = latestVersion
	}

	for _, stagedProduct := range diagnosticReport.StagedProducts {
		if stagedProduct.Name == productName && stagedProduct.Version == productVersion {
			sp.logger.Printf("%s %s is already staged", productName, productVersion)
			return nil
		}
	}

	available, err := sp.service.CheckProductAvailability(productName, productVersion)
	if err != nil {
		return fmt.Errorf("failed to stage product: cannot check availability of product %s %s", productName, productVersion)
	}

	if !available {
		return fmt.Errorf("failed to stage product: cannot find product %s %s", productName, productVersion)
	}

	sp.logger.Printf("staging %s %s", productName, productVersion)

	err = sp.service.Stage(api.StageProductInput{
		ProductName:    productName,
		ProductVersion: productVersion,
	}, deployedProductGUID)
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	sp.logger.Printf("finished staging")

	return nil
}

func (sp *StageProduct) loadConfig() error {
	if sp.Options.ConfigFile != "" {
		contents, err := os.ReadFile(sp.Options.ConfigFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(contents, &sp.Options)
		if err != nil {
			return err
		}
	}

	productName := sp.Options.Product
	productVersion := sp.Options.Version
	if productName == "" || productVersion == "" {
		return fmt.Errorf("--product-name (%s) and --product-version (%s) are required", productName, productVersion)
	}

	return nil
}
