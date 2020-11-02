package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DeployedProducts struct {
	presenter presenters.FormattedPresenter
	service   deployedProductsService
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//counterfeiter:generate -o ./fakes/deployed_products_service.go --fake-name DeployedProductsService . deployedProductsService
type deployedProductsService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewDeployedProducts(presenter presenters.FormattedPresenter, service deployedProductsService) *DeployedProducts {
	return &DeployedProducts{
		presenter: presenter,
		service:   service,
	}
}

func (dp DeployedProducts) Execute(args []string) error {
	diagnosticReport, err := dp.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve deployed products %s", err)
	}

	deployedProducts := diagnosticReport.DeployedProducts

	dp.presenter.SetFormat(dp.Options.Format)
	dp.presenter.PresentDeployedProducts(deployedProducts)

	return nil
}
