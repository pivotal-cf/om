package commands

import (
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

//counterfeiter:generate -o ./fakes/available_products_service.go --fake-name AvailableProductsService . availableProductsService

type availableProductsService interface {
	ListAvailableProducts() (api.AvailableProductsOutput, error)
}

func NewAvailableProducts(apService availableProductsService, presenter presenters.FormattedPresenter, logger logger) *AvailableProducts {
	return &AvailableProducts{service: apService, presenter: presenter, logger: logger}
}

func (ap AvailableProducts) Execute(args []string) error {
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
