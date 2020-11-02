package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
)

type StagedManifest struct {
	service stagedManifestService
	logger  logger
	Options struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

//counterfeiter:generate -o ./fakes/staged_manifest_service.go --fake-name StagedManifestService . stagedManifestService
type stagedManifestService interface {
	GetStagedProductByName(product string) (api.StagedProductsFindOutput, error)
	GetStagedProductManifest(guid string) (string, error)
}

func NewStagedManifest(service stagedManifestService, logger logger) *StagedManifest {
	return &StagedManifest{
		service: service,
		logger:  logger,
	}
}

func (sm StagedManifest) Execute(args []string) error {
	output, err := sm.service.GetStagedProductByName(sm.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find product: %s", err)
	}

	manifest, err := sm.service.GetStagedProductManifest(output.Product.GUID)
	if err != nil {
		return fmt.Errorf("failed to fetch product manifest: %s", err)
	}

	sm.logger.Print(manifest)

	return nil
}
