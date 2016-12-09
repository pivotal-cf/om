package commands

import "fmt"

type DeleteUnusedProducts struct {
	productsService productUploader
	logger          logger
}

func NewDeleteUnusedProducts(productUploader productUploader, logger logger) DeleteUnusedProducts {
	return DeleteUnusedProducts{
		productsService: productUploader,
		logger:          logger,
	}
}

func (dup DeleteUnusedProducts) Usage() Usage {
	return Usage{
		Description:      "This command deletes unused products in the targeted Ops Manager",
		ShortDescription: "deletes unused products on the Ops Manager targeted",
	}
}

func (dup DeleteUnusedProducts) Execute(args []string) error {
	dup.logger.Printf("trashing unused products")

	err := dup.productsService.Trash()
	if err != nil {
		return fmt.Errorf("could not delete unused products: %s", err)
	}

	dup.logger.Printf("done")

	return nil
}
