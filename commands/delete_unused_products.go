package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeleteUnusedProducts struct {
	productsService ps
	logger          logger
}

func NewDeleteUnusedProducts(productDeleter ps, logger logger) DeleteUnusedProducts {
	return DeleteUnusedProducts{
		productsService: productDeleter,
		logger:          logger,
	}
}

func (dup DeleteUnusedProducts) Execute(args []string) error {
	dup.logger.Printf("trashing unused products")

	err := dup.productsService.Delete(api.AvailableProductsInput{}, true)
	if err != nil {
		return err
	}

	dup.logger.Printf("done")

	return nil
}

func (dup DeleteUnusedProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command deletes unused products in the targeted Ops Manager",
		ShortDescription: "deletes unused products on the Ops Manager targeted",
	}
}
