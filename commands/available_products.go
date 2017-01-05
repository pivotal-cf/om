package commands

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf/om/api"
)

type AvailableProducts struct {
	service     availableProductsService
	tableWriter tableWriter
	logger      logger
}

//go:generate counterfeiter -o ./fakes/available_products_service.go --fake-name AvailableProductsService . availableProductsService
type availableProductsService interface {
	List() (api.AvailableProductsOutput, error)
}

//go:generate counterfeiter -o ./fakes/table_writer.go --fake-name TableWriter . tableWriter
type tableWriter interface {
	SetHeader([]string)
	Append([]string)
	SetAlignment(int)
	Render()
}

func NewAvailableProducts(apService availableProductsService, tableWriter tableWriter, logger logger) AvailableProducts {
	return AvailableProducts{service: apService, tableWriter: tableWriter, logger: logger}
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

	ap.tableWriter.SetAlignment(tablewriter.ALIGN_LEFT)
	ap.tableWriter.SetHeader([]string{"Name", "Version"})

	for _, product := range output.ProductsList {
		ap.tableWriter.Append([]string{product.Name, product.Version})
	}

	ap.tableWriter.Render()

	return nil
}

func (ap AvailableProducts) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all available products.",
		ShortDescription: "list available products",
	}
}
