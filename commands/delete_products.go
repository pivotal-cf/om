package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeleteProduct struct {
	service deleteProductService
	Options struct {
		Product string `long:"product-name"    short:"p" required:"true" description:"name of product"`
		Version string `long:"product-version" short:"v" required:"true" description:"version of product"`
	}
}

//go:generate counterfeiter -o ./fakes/delete_product_service.go --fake-name DeleteProductService . deleteProductService
type deleteProductService interface {
	DeleteAvailableProducts(input api.DeleteAvailableProductsInput) error
}

func NewDeleteProduct(service deleteProductService) DeleteProduct {
	return DeleteProduct{
		service: service,
	}
}

func (dp DeleteProduct) Execute(args []string) error {
	if _, err := jhanda.Parse(&dp.Options, args); err != nil {
		return fmt.Errorf("could not parse delete-product flags: %s", err)
	}

	err := dp.service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
		ProductName:             dp.Options.Product,
		ProductVersion:          dp.Options.Version,
		ShouldDeleteAllProducts: false,
	})
	if err != nil {
		return err
	}

	return nil
}

func (dp DeleteProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command deletes the named product from the targeted Ops Manager",
		ShortDescription: "deletes a product from the Ops Manager",
		Flags:            dp.Options,
	}
}
