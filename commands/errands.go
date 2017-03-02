package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

//go:generate counterfeiter -o ./fakes/staged_products_finder.go --fake-name StagedProductsFinder . stagedProductsFinder
type stagedProductsFinder interface {
	Find(productName string) (api.StagedProductsFindOutput, error)
}

//go:generate counterfeiter -o ./fakes/errands_service.go --fake-name ErrandsService . errandsService
type errandsService interface {
	List(productID string) (api.ErrandsListOutput, error)
}

type Errands struct {
	tableWriter          tableWriter
	errandsService       errandsService
	stagedProductsFinder stagedProductsFinder
	Options              struct {
		ProductName string `short:"p" long:"product-name" description:"name of product"`
	}
}

func NewErrands(tableWriter tableWriter, errandsService errandsService, stagedProductsFinder stagedProductsFinder) Errands {
	return Errands{
		tableWriter:          tableWriter,
		errandsService:       errandsService,
		stagedProductsFinder: stagedProductsFinder,
	}
}

func (e Errands) Execute(args []string) error {
	_, err := flags.Parse(&e.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse errands flags: %s", err)
	}

	if e.Options.ProductName == "" {
		return errors.New("error: product-name is missing. Please see usage for more information.")
	}

	findOutput, err := e.stagedProductsFinder.Find(e.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find staged product %q: %s", e.Options.ProductName, err)
	}

	errandsOutput, err := e.errandsService.List(findOutput.Product.GUID)
	if err != nil {
		return fmt.Errorf("failed to list errands: %s", err)
	}

	e.tableWriter.SetHeader([]string{"Name", "Post Deploy Enabled"})
	for _, errand := range errandsOutput.Errands {
		e.tableWriter.Append([]string{errand.Name, strconv.FormatBool(errand.PostDeploy)})
	}

	e.tableWriter.Render()

	return nil
}

func (e Errands) Usage() Usage {
	return Usage{
		Description:      "This authenticated command lists all errands for a product.",
		ShortDescription: "list errands for a product",
	}
}
