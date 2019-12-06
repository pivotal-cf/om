package commands

import (
	"fmt"
	"github.com/pivotal-cf/om/api"

	"github.com/pivotal-cf/jhanda"
)

type ProductDiff struct {
	service productDiffService
	logger  logger
	Options struct {
		Product string `long:"product" short:"p" description:"Product to get diff for"`
	}
}

//counterfeiter:generate -o ./fakes/diff_service.go --fake-name ProductDiffService . productDiffService
type productDiffService interface {
	ProductDiff(productName string) (api.ProductDiff, error)
}

func NewProductDiff(service productDiffService, logger logger) ProductDiff {
	return ProductDiff{
		service: service,
		logger:  logger,
	}
}

func (c ProductDiff) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse product-diff flags: %s", err)
	}
	diff, err := c.service.ProductDiff(c.Options.Product)
	if err != nil {
		return err
	}
	c.logger.Printf("Status: %s\n%s", diff.Manifest.Status, diff.Manifest.Diff)

	return nil
}

func (c ProductDiff) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command displays the bosh manifest diff for a product (Note: property values are redacted and will appear as '***')",
		ShortDescription: "displays BOSH manifest diff for a product",
		Flags:            c.Options,
	}
}
