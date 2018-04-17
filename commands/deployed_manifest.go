package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeployedManifest struct {
	service deployedManifestService
	logger  logger
	Options struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

//go:generate counterfeiter -o ./fakes/deployed_manifest_service.go --fake-name DeployedManifestService . deployedManifestService
type deployedManifestService interface {
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	GetDeployedProductManifest(guid string) (string, error)
}

func NewDeployedManifest(service deployedManifestService, logger logger) DeployedManifest {
	return DeployedManifest{
		service: service,
		logger:  logger,
	}
}

func (dm DeployedManifest) Execute(args []string) error {
	if _, err := jhanda.Parse(&dm.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-manifest flags: %s", err)
	}

	output, err := dm.service.ListDeployedProducts()
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

	manifest, err := dm.service.GetDeployedProductManifest(guid)
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
