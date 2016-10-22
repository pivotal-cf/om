package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/common"
	"github.com/pivotal-cf/om/flags"
)

type StageProduct struct {
	logger         common.Logger
	productService productService
	Options        struct {
		Product string `short:"p"  long:"product-name"  description:"name of product"`
		Version string `short:"v"  long:"product-version"  description:"version of product"`
	}
}

func NewStageProduct(productService productService, logger common.Logger) StageProduct {
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

	if sp.Options.Product == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	if sp.Options.Version == "" {
		return errors.New("error: product-version is missing. Please see usage for more information.")
	}

	sp.logger.Printf("staging %s %s", sp.Options.Product, sp.Options.Version)

	err = sp.productService.Stage(api.StageProductInput{
		ProductName:    sp.Options.Product,
		ProductVersion: sp.Options.Version,
	})
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	sp.logger.Printf("finished staging")

	return nil
}
