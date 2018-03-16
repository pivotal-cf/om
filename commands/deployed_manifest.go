package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type DeployedManifest struct {
	logger           logger
	deployedProducts deployedProductsLister
	Options          struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

func NewDeployedManifest(logger logger, deployedProducts deployedProductsLister) DeployedManifest {
	return DeployedManifest{
		logger:           logger,
		deployedProducts: deployedProducts,
	}
}

func (dm DeployedManifest) Execute(args []string) error {
	if _, err := jhanda.Parse(&dm.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-manifest flags: %s", err)
	}

	output, err := dm.deployedProducts.List()
	if err != nil {
		return err
	}

	var guid string
	for _, product := range output {
		if product.Type == dm.Options.ProductName {
			guid = product.GUID
			break
		}
	}

	if guid == "" {
		return errors.New("could not find given product")
	}

	manifest, err := dm.deployedProducts.Manifest(guid)
	if err != nil {
		return err
	}

	dm.logger.Print(manifest)

	return nil
}

func (dm DeployedManifest) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command prints the deployed manifest for a product",
		ShortDescription: "prints the deployed manifest for a product",
		Flags:            dm.Options,
	}
}
