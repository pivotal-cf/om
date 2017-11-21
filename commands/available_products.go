package commands

import (
	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

type AvailableProducts struct {
	service   availableProductsService
	presenter presenter
	logger    logger
}

//go:generate counterfeiter -o ./fakes/available_products_service.go --fake-name AvailableProductsService . availableProductsService

type availableProductsService interface {
	List() (api.AvailableProductsOutput, error)
}

func NewAvailableProducts(apService availableProductsService, presenter presenter, logger logger) AvailableProducts {
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

	products := []models.Product{}
	for _, product := range output.ProductsList {
		products = append(products, models.Product{
			Name:    product.Name,
			Version: product.Version,
		})
	}

	ap.presenter.PresentAvailableProducts(products)

	return nil
}

func (ap AvailableProducts) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command lists all available products.",
		ShortDescription: "list available products",
	}
}
