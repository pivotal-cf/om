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
		Product string `long:"product" short:"p" required:"true" description:"Product to get diff for"`
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

	c.logger.Println("## Product Manifest\n")
	if diff.Manifest.Status == "same" {
		c.logger.Printf("no changes")
		c.printRuntimeConfigs(diff)
		return nil
	}

	c.logger.Printf("%s\n\n", diff.Manifest.Diff)
	c.printRuntimeConfigs(diff)

	return nil
}

func (c ProductDiff) printRuntimeConfigs(diff api.ProductDiff) {
	c.logger.Println("## Runtime Configs\n")

	noneChanged := true

	for _, config := range diff.RuntimeConfigs {
		if config.Status == "same" {
			continue
		}

		noneChanged = false
		c.logger.Printf("### %s\n\n", config.Name)
		c.logger.Printf("%s\n\n", config.Diff)
	}

	if noneChanged {
		c.logger.Printf("no changes")
	}
}

func (c ProductDiff) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "**EXPERIMENTAL** This command displays the bosh manifest diff for a product (Note: secret values are replaced with double-paren variable names)",
		ShortDescription: "**EXPERIMENTAL** displays BOSH manifest diff for a product",
		Flags:            c.Options,
	}
}
