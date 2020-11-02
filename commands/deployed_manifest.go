package commands

import (
	"errors"
	"github.com/pivotal-cf/om/api"
)

type DeployedManifest struct {
	service deployedManifestService
	logger  logger
	Options struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

//counterfeiter:generate -o ./fakes/deployed_manifest_service.go --fake-name DeployedManifestService . deployedManifestService
type deployedManifestService interface {
	ListDeployedProducts() ([]api.DeployedProductOutput, error)
	GetDeployedProductManifest(guid string) (string, error)
}

func NewDeployedManifest(service deployedManifestService, logger logger) *DeployedManifest {
	return &DeployedManifest{
		service: service,
		logger:  logger,
	}
}

func (dm DeployedManifest) Execute(args []string) error {
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
