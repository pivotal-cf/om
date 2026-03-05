package commands

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
)

type StageProduct struct {
	logger  logger
	service stageProductService
	Options struct {
		ConfigFile       string `long:"config" description:"the config file to load product name and version (can be same as the product configuration file)"`
		Product          string `yaml:"product-name" long:"product-name"    short:"p" description:"name of product"`
		Version          string `yaml:"product-version" long:"product-version"        description:"version of product"`
		StageAllReplicas bool   `yaml:"stage-all-replicas" long:"stage-all-replicas" description:"stage this product for all replicas of this product, default false"`
		StageReplicas    string `yaml:"stage-replicas" long:"stage-replicas" description:"accepts a comma-separated list of tile names of a replicated product to stage"`
	}
}

//counterfeiter:generate -o ./fakes/stage_product_service.go --fake-name StageProductService . stageProductService
type stageProductService interface {
	CheckProductAvailability(productName string, productVersion string) (bool, error)
	GetDiagnosticReport() (api.DiagnosticReport, error)
	GetLatestAvailableVersion(productName string) (string, error)
	Info() (api.Info, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	ListStagedProducts() (api.StagedProductsOutput, error)
	Stage(api.StageProductInput, string) error
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
	replicaFlagsUsed := sp.Options.StageAllReplicas || sp.Options.StageReplicas != ""

	if err := sp.validateReplicaFlags(); err != nil {
		return err
	}

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

	if replicaFlagsUsed {
		return sp.stageReplicas(productName, productVersion)
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

func (sp StageProduct) stageReplicas(productName, productVersion string) error {
	type replicaInfo struct {
		Type string
		GUID string
	}

	var replicas []replicaInfo

	stagedProducts, err := sp.service.ListStagedProducts()
	if err != nil {
		return fmt.Errorf("failed to list staged products: %s", err)
	}
	includeAll := sp.Options.StageAllReplicas

	for _, p := range stagedProducts.Products {
		templateName := p.ProductTemplateName
		if templateName == "" {
			templateName = p.Type
		}
		if templateName == productName && (includeAll || p.Type != productName) {
			replicas = append(replicas, replicaInfo{Type: p.Type, GUID: p.GUID})
		}
	}

	if sp.Options.StageReplicas != "" {
		requestedNames := strings.Split(sp.Options.StageReplicas, ",")
		requestedSet := make(map[string]bool)
		for _, name := range requestedNames {
			requestedSet[strings.TrimSpace(name)] = true
		}

		var filtered []replicaInfo
		for _, r := range replicas {
			if requestedSet[r.Type] {
				filtered = append(filtered, r)
				delete(requestedSet, r.Type)
			}
		}

		if len(requestedSet) > 0 {
			var missing []string
			for name := range requestedSet {
				missing = append(missing, name)
			}
			return fmt.Errorf("failed to stage replicas: could not find replicas with type(s): %s", strings.Join(missing, ", "))
		}

		replicas = filtered
	}

	for _, replica := range replicas {
		sp.logger.Printf("staging replica %s %s", replica.Type, productVersion)
		err := sp.service.Stage(api.StageProductInput{
			ProductName:    productName,
			ProductVersion: productVersion,
		}, replica.GUID)
		if err != nil {
			return fmt.Errorf("failed to stage replica %s: %s", replica.Type, err)
		}
	}

	if len(replicas) > 0 {
		sp.logger.Printf("finished staging replicas")
	}

	return nil
}

func (sp StageProduct) validateReplicaFlags() error {
	if sp.Options.StageAllReplicas && sp.Options.StageReplicas != "" {
		return fmt.Errorf("--stage-all-replicas and --stage-replicas are mutually exclusive")
	}

	if sp.Options.StageAllReplicas || sp.Options.StageReplicas != "" {
		info, err := sp.service.Info()
		if err != nil {
			return fmt.Errorf("failed to get Ops Manager version: %s", err)
		}
		if ok, verErr := info.VersionAtLeast(3, 3); !ok {
			if verErr != nil {
				return fmt.Errorf("stage-product replica flags require Ops Manager 3.3 or newer: %w", verErr)
			}
			return fmt.Errorf("stage-product replica flags require Ops Manager 3.3 or newer (current version: %s)", info.Version)
		}
	}

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
