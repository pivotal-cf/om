package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type ConfigureProduct struct {
	productsService productConfigurer
	logger          logger
	Options         struct {
		ProductName       string `short:"n"  long:"product-name" description:"name of the product being configured"`
		ProductProperties string `short:"p" long:"product-properties" description:"properties to be configured in JSON format" default:"{}"`
		NetworkProperties string `short:"pn" long:"product-network" descriptions:"properties to be configured in JSON format" default:"{}"`
	}
}

//go:generate counterfeiter -o ./fakes/product_configurer.go --fake-name ProductConfigurer . productConfigurer
type productConfigurer interface {
	StagedProducts() (api.StagedProductsOutput, error)
	Configure(api.ProductsConfigurationInput) error
}

func NewConfigureProduct(productConfigurer productConfigurer, logger logger) ConfigureProduct {
	return ConfigureProduct{
		productsService: productConfigurer,
		logger:          logger,
	}
}

func (cp ConfigureProduct) Execute(args []string) error {
	_, err := flags.Parse(&cp.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse configure-product flags: %s", err)
	}

	if cp.Options.ProductName == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	if cp.Options.ProductProperties == "{}" && cp.Options.NetworkProperties == "{}" {
		cp.logger.Printf("Provided properties are empty, nothing to do here")
		return nil
	}

	cp.logger.Printf("setting properties")
	stagedProducts, err := cp.productsService.StagedProducts()
	if err != nil {
		return err
	}

	var productGUID string
	for _, sp := range stagedProducts.Products {
		if sp.Type == cp.Options.ProductName {
			productGUID = sp.GUID
			break
		}
	}

	err = cp.productsService.Configure(api.ProductsConfigurationInput{
		GUID:          productGUID,
		Configuration: cp.Options.ProductProperties,
		Network:       cp.Options.NetworkProperties,
	})
	if err != nil {
		return fmt.Errorf("failed to configure product: %s", err)
	}

	cp.logger.Printf("finished setting properties")

	return nil
}

func (cp ConfigureProduct) Usage() Usage {
	return Usage{
		Description:      "This authenticated command configures a staged product",
		ShortDescription: "configures a staged product",
		Flags:            cp.Options,
	}
}
