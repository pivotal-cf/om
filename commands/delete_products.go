package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type DeleteProduct struct {
	productsService ps
	Options         struct {
		Product string `short:"p"  long:"product-name"  description:"name of product"`
		Version string `short:"v"  long:"product-version"  description:"version of product"`
	}
}

//go:generate counterfeiter -o ./fakes/product_deleter.go --fake-name ProductDeleter . ps
type ps interface {
	Delete(input api.AvailableProductsInput, deleteAll bool) error
}

func NewDeleteProduct(productsService ps) DeleteProduct {
	return DeleteProduct{
		productsService: productsService,
	}
}

func (dp DeleteProduct) Execute(args []string) error {
	_, err := flags.Parse(&dp.Options, args)
	if err != nil {
		return err
	}

	err = dp.productsService.Delete(api.AvailableProductsInput{
		ProductName:    dp.Options.Product,
		ProductVersion: dp.Options.Version,
	}, false)
	if err != nil {
		return err
	}

	return nil
}

func (dp DeleteProduct) Usage() Usage {
	return Usage{
		Description:      "This command deletes the named product from the targeted Ops Manager",
		ShortDescription: "deletes a product from the Ops Manager",
		Flags:            dp.Options,
	}
}
