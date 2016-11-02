package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type StageProduct struct {
	logger          logger
	productsService productStager
	Options         struct {
		Product string `short:"p"  long:"product-name"  description:"name of product"`
		Version string `short:"v"  long:"product-version"  description:"version of product"`
	}
}

//go:generate counterfeiter -o ./fakes/product_stager.go --fake-name ProductStager . productStager
type productStager interface {
	Stage(api.StageProductInput) error
}

func NewStageProduct(productStager productStager, logger logger) StageProduct {
	return StageProduct{
		logger:          logger,
		productsService: productStager,
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

	err = sp.productsService.Stage(api.StageProductInput{
		ProductName:    sp.Options.Product,
		ProductVersion: sp.Options.Version,
	})
	if err != nil {
		return fmt.Errorf("failed to stage product: %s", err)
	}

	sp.logger.Printf("finished staging")

	return nil
}
