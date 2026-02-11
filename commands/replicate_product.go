package commands

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf/om/api"
)

type ReplicateProduct struct {
	logger  logger
	service stageProductService
	Options struct {
		ConfigFile    string `long:"config" description:"the config file to load product name and version"`
		Product       string `yaml:"product-name" long:"product-name" short:"p" description:"name of product"`
		Version       string `yaml:"product-version" long:"product-version" description:"version of product"`
		ReplicaSuffix string `yaml:"replica-suffix" long:"replica-suffix" description:"suffix for the new staged product"`
	}
}

func NewReplicateProduct(service stageProductService, logger logger) *ReplicateProduct {
	return &ReplicateProduct{
		logger:  logger,
		service: service,
	}
}

func (rp ReplicateProduct) Execute(args []string) error {
	err := rp.loadConfig()
	if err != nil {
		return err
	}

	productName := rp.Options.Product
	productVersion := rp.Options.Version
	replicaSuffix := rp.Options.ReplicaSuffix

	err = checkRunningInstallation(rp.service.ListInstallations)
	if err != nil {
		return err
	}

	diagnosticReport, err := rp.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to replicate product: %s", err)
	}

	replicaType := productName + "-" + replicaSuffix
	for _, stagedProduct := range diagnosticReport.StagedProducts {
		if stagedProduct.Name == replicaType && stagedProduct.Version == productVersion {
			rp.logger.Printf("%s %s with suffix %s is already staged", productName, productVersion, replicaSuffix)
			return nil
		}
	}

	if productVersion == "latest" {
		latestVersion, err := rp.service.GetLatestAvailableVersion(productName)
		if err != nil {
			return fmt.Errorf("could not find latest version: %w", err)
		}
		productVersion = latestVersion
	}

	available, err := rp.service.CheckProductAvailability(productName, productVersion)
	if err != nil {
		return fmt.Errorf("failed to replicate product: cannot check availability of product %s %s", productName, productVersion)
	}

	if !available {
		return fmt.Errorf("failed to replicate product: cannot find product %s %s", productName, productVersion)
	}

	rp.logger.Printf("replicating %s %s with suffix %s", productName, productVersion, replicaSuffix)

	err = rp.service.Stage(api.StageProductInput{
		ProductName:    productName,
		ProductVersion: productVersion,
		Replicate:      true,
		ReplicaSuffix:  replicaSuffix,
	}, "")
	if err != nil {
		return fmt.Errorf("failed to replicate product: %s", err)
	}

	rp.logger.Printf("finished replicating")

	return nil
}

func (rp *ReplicateProduct) loadConfig() error {
	if rp.Options.ConfigFile != "" {
		contents, err := os.ReadFile(rp.Options.ConfigFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(contents, &rp.Options)
		if err != nil {
			return err
		}
	}

	productName := rp.Options.Product
	productVersion := rp.Options.Version
	replicaSuffix := rp.Options.ReplicaSuffix
	if productName == "" || productVersion == "" || replicaSuffix == "" {
		return fmt.Errorf("--product-name (%s), --product-version (%s), and --replica-suffix (%s) are required", productName, productVersion, replicaSuffix)
	}

	return nil
}
