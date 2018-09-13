package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

type AvailableProducts struct {
	service   availableProductsService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/available_products_service.go --fake-name AvailableProductsService . availableProductsService

type availableProductsService interface {
	ListAvailableProducts() (api.AvailableProductsOutput, error)
}

func NewAvailableProducts(apService availableProductsService, presenter presenters.FormattedPresenter, logger logger) AvailableProducts {
	return AvailableProducts{service: apService, presenter: presenter, logger: logger}
}

func (ap AvailableProducts) Execute(args []string) error {
	if _, err := jhanda.Parse(&ap.Options, args); err != nil {
		return fmt.Errorf("could not parse available-products flags: %s", err)
	}

	output, err := ap.service.ListAvailableProducts()
	if err != nil {
		return err
	}

	if len(output.ProductsList) == 0 {
		ap.logger.Printf("no available products found")
		return nil
	}

	var products []models.Product
	for _, product := range output.ProductsList {
		products = append(products, models.Product{
			Name:    product.Name,
			Version: product.Version,
		})
	}

	ap.presenter.SetFormat(ap.Options.Format)
	ap.presenter.PresentAvailableProducts(products)

	return nil
}

func (ap AvailableProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all available products.",
		ShortDescription: "list available products",
		Flags:            ap.Options,
	}
}
