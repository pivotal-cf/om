package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/metadata"
	"os"
)

type CheckProductUploaded struct {
	service   availableProductsService

	Options struct {
		interpolateConfigFileOptions
		PivnetOptions
	}
}

func (c CheckProductUploaded) Execute(args []string) error {
	err := loadConfigFile(args, &c.Options, os.Environ)
	if err != nil {
		return fmt.Errorf("could not parse check-product-uploaded flags: %s", err)
	}

	options := c.Options
	provider := metadata.NewPivnetProvider(
		options.PivnetHost,
		options.PivnetToken,
		options.PivnetProductSlug,
		options.ProductVersion,
		options.FileGlob,
		options.PivnetDisableSSL,
	)
	contents, err := provider.MetadataBytes()
	if err != nil {
		return fmt.Errorf("could not parse response from Pivnet: %s", err)
	}

	metadata, err := generator.NewMetadata(contents)
	if err != nil {
		return fmt.Errorf("could not parse metadata from Pivnet product (please contact Pivotal support): %s", err)
	}

	name, version := metadata.ProductName(), metadata.ProductVersion()

	output, err := c.service.ListAvailableProducts()
	if err != nil {
		return fmt.Errorf("could not find products on OpsManager: %s", err)
	}

	for _, productInfo := range output.ProductsList {
		if productInfo.Name == name && productInfo.Version == version {
			return nil
		}
	}

	return fmt.Errorf("product %s of version %s could not be found on OpsManager", name, version)
}

func (c CheckProductUploaded) Usage() jhanda.Usage {
	return jhanda.Usage{}
}

func NewCheckProductUploaded(apService availableProductsService) *CheckProductUploaded {
	return &CheckProductUploaded{
		service: apService,
	}
}
