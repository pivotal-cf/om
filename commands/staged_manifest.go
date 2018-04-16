package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type StagedManifest struct {
	stagedProductsService stagedProductsService
	logger                logger
	Options               struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

func NewStagedManifest(stagedProductsService stagedProductsService, logger logger) StagedManifest {
	return StagedManifest{
		stagedProductsService: stagedProductsService,
		logger:                logger,
	}
}

func (sm StagedManifest) Execute(args []string) error {
	if _, err := jhanda.Parse(&sm.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-manifest flags: %s", err)
	}

	output, err := sm.stagedProductsService.Find(sm.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find product: %s", err)
	}

	manifest, err := sm.stagedProductsService.GetStagedProductManifest(output.Product.GUID)
	if err != nil {
		return fmt.Errorf("failed to fetch product manifest: %s", err)
	}

	sm.logger.Print(manifest)

	return nil
}

func (sm StagedManifest) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command prints the staged manifest for a product",
		ShortDescription: "prints the staged manifest for a product",
		Flags:            sm.Options,
	}
}
