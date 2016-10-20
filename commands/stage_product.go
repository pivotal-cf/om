package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type StageProduct struct {
	logger         logger
	productService productService
	Options        struct {
		Product string `short:"p"  long:"product"  description:"name of product"`
	}
}

func NewStageProduct(productService productService, logger logger) StageProduct {
	return StageProduct{
		logger:         logger,
		productService: productService,
	}
}

func (sp StageProduct) Usage() Usage {
	return Usage{
		Description:      "This command attempts to stage a product in the Ops Manager",
		ShortDescription: "stages a given product in the Ops Manager targeted",
		Flags:            sp.Options,
	}
}

func (sp StageProduct) Execute(args []string) error {
	_, err := flags.Parse(&sp.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse stage-product flags: %s", err)
	}

	sp.logger.Printf("staging %s", sp.Options.Product)

	err = sp.productService.Stage(api.StageProductInput{
		ProductName: sp.Options.Product,
	})
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	sp.logger.Printf("finished staging")

	return nil
}
