package commands

import (
	"github.com/pivotal-cf/om/api"
)

type DeleteProduct struct {
	service deleteProductService
	Options struct {
		Product string `long:"product-name"    short:"p" required:"true" description:"name of product"`
		Version string `long:"product-version" short:"v" required:"true" description:"version of product"`
	}
}

//counterfeiter:generate -o ./fakes/delete_product_service.go --fake-name DeleteProductService . deleteProductService
type deleteProductService interface {
	DeleteAvailableProducts(input api.DeleteAvailableProductsInput) error
}

func NewDeleteProduct(service deleteProductService) *DeleteProduct {
	return &DeleteProduct{
		service: service,
	}
}

func (dp DeleteProduct) Execute(args []string) error {
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
