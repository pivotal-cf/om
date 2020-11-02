package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
)

type UnstageProduct struct {
	logger  logger
	service unstageProductService
	Options struct {
		Product string `long:"product-name" short:"p" required:"true" description:"name of product"`
	}
}

//counterfeiter:generate -o ./fakes/unstage_product_service.go --fake-name UnstageProductService . unstageProductService
type unstageProductService interface {
	DeleteStagedProduct(api.UnstageProductInput) error
}

func NewUnstageProduct(service unstageProductService, logger logger) *UnstageProduct {
	return &UnstageProduct{
		logger:  logger,
		service: service,
	}
}

func (up UnstageProduct) Execute(args []string) error {
	up.logger.Printf("unstaging %s", up.Options.Product)

	err := up.service.DeleteStagedProduct(api.UnstageProductInput{
		ProductName: up.Options.Product,
	})

	if err != nil {
		return fmt.Errorf("failed to unstage product: %s", err)
	}

	up.logger.Printf("finished unstaging")

	return nil
}
