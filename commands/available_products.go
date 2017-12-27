package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

type AvailableProducts struct {
	service   availableProductsService
	presenter presenters.Presenter
	logger    logger
}

//go:generate counterfeiter -o ./fakes/available_products_service.go --fake-name AvailableProductsService . availableProductsService

type availableProductsService interface {
	List() (api.AvailableProductsOutput, error)
}

func NewAvailableProducts(apService availableProductsService, presenter presenters.Presenter, logger logger) AvailableProducts {
	return AvailableProducts{service: apService, presenter: presenter, logger: logger}
}

func (ap AvailableProducts) Execute(args []string) error {
	output, err := ap.service.List()
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

	ap.presenter.PresentAvailableProducts(products)

	return nil
}

func (ap AvailableProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all available products.",
		ShortDescription: "list available products",
	}
}
