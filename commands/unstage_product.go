package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type UnstageProduct struct {
	logger                logger
	stagedProductsService productUnstager
	Options               struct {
		Product string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

//go:generate counterfeiter -o ./fakes/product_unstager.go --fake-name ProductUnstager . productUnstager
type productUnstager interface {
	Unstage(api.UnstageProductInput) error
}

func NewUnstageProduct(productUnstager productUnstager, logger logger) UnstageProduct {
	return UnstageProduct{
		logger:                logger,
		stagedProductsService: productUnstager,
	}
}

func (up UnstageProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&up.Options, args); err != nil {
		return fmt.Errorf("could not parse unstage-product flags: %s", err)
	}

	up.logger.Printf("unstaging %s", up.Options.Product)

	err := up.stagedProductsService.Unstage(api.UnstageProductInput{
		ProductName: up.Options.Product,
	})

	if err != nil {
		return fmt.Errorf("failed to unstage product: %s", err)
	}

	up.logger.Printf("finished unstaging")

	return nil
}

func (up UnstageProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to unstage a product from the Ops Manager",
		ShortDescription: "unstages a given product from the Ops Manager targeted",
		Flags:            up.Options,
	}
}
