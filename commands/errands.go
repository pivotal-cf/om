package commands

import (
	"fmt"
	"strconv"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
	"github.com/pivotal-cf/om/presenters"
)

//go:generate counterfeiter -o ./fakes/errands_service.go --fake-name ErrandsService . errandsService
type errandsService interface {
	GetStagedProductByName(productName string) (api.StagedProductsFindOutput, error)
	ListStagedProductErrands(productID string) (api.ErrandsListOutput, error)
}

type Errands struct {
	presenter presenters.FormattedPresenter
	service   errandsService
	Options   struct {
		ProductName string `long:"product-name" short:"p" required:"true" description:"name of product"`
		Format      string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewErrands(presenter presenters.FormattedPresenter, service errandsService) Errands {
	return Errands{
		presenter: presenter,
		service:   service,
	}
}

func (e Errands) Execute(args []string) error {
	if _, err := jhanda.Parse(&e.Options, args); err != nil {
		return fmt.Errorf("could not parse errands flags: %s", err)
	}

	findOutput, err := e.service.GetStagedProductByName(e.Options.ProductName)
	if err != nil {
		return fmt.Errorf("failed to find staged product %q: %s", e.Options.ProductName, err)
	}

	errandsOutput, err := e.service.ListStagedProductErrands(findOutput.Product.GUID)
	if err != nil {
		return fmt.Errorf("failed to list errands: %s", err)
	}

	var errands []models.Errand
	for _, errand := range errandsOutput.Errands {
		errands = append(errands, models.Errand{
			Name:              errand.Name,
			PostDeployEnabled: boolStringFromType(errand.PostDeploy),
			PreDeleteEnabled:  boolStringFromType(errand.PreDelete),
		})
	}

	e.presenter.SetFormat(e.Options.Format)
	e.presenter.PresentErrands(errands)

	return nil
}

func boolStringFromType(object interface{}) string {
	switch p := object.(type) {
	case string:
		return p
	case bool:
		return strconv.FormatBool(p)
	default:
		return ""
	}
}

func (e Errands) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all errands for a product.",
		ShortDescription: "list errands for a product",
		Flags:            e.Options,
	}
}
