package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type StageProduct struct {
	logger  logger
	service stageProductService
	Options struct {
		Product string `long:"product-name"    short:"p" required:"true" description:"name of product"`
		Version string `long:"product-version" short:"v" required:"true" description:"version of product"`
	}
}

//go:generate counterfeiter -o ./fakes/stage_product_service.go --fake-name StageProductService . stageProductService
type stageProductService interface {
	CheckProductAvailability(productName string, productVersion string) (bool, error)
	GetDiagnosticReport() (api.DiagnosticReport, error)
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	Stage(api.StageProductInput, string) error
}

func NewStageProduct(service stageProductService, logger logger) StageProduct {
	return StageProduct{
		logger:  logger,
		service: service,
	}
}

func (sp StageProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&sp.Options, args); err != nil {
		return fmt.Errorf("could not parse stage-product flags: %s", err)
	}

	err := checkRunningInstallation(sp.service.ListInstallations)
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
		if deployedProduct.Type == sp.Options.Product {
			deployedProductGUID = deployedProduct.GUID
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	for _, stagedProduct := range diagnosticReport.StagedProducts {
		if stagedProduct.Name == sp.Options.Product && stagedProduct.Version == sp.Options.Version {
			sp.logger.Printf("%s %s is already staged", sp.Options.Product, sp.Options.Version)
			return nil
		}
	}

	available, err := sp.service.CheckProductAvailability(sp.Options.Product, sp.Options.Version)
	if err != nil {
		return fmt.Errorf("failed to stage product: cannot check availability of product %s %s", sp.Options.Product, sp.Options.Version)
	}

	if !available {
		return fmt.Errorf("failed to stage product: cannot find product %s %s", sp.Options.Product, sp.Options.Version)
	}

	sp.logger.Printf("staging %s %s", sp.Options.Product, sp.Options.Version)

	err = sp.service.Stage(api.StageProductInput{
		ProductName:    sp.Options.Product,
		ProductVersion: sp.Options.Version,
	}, deployedProductGUID)
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	sp.logger.Printf("finished staging")

	return nil
}

func (sp StageProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to stage a product in the Ops Manager",
		ShortDescription: "stages a given product in the Ops Manager targeted",
		Flags:            sp.Options,
	}
}
