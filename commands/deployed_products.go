package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DeployedProducts struct {
	presenter presenters.Presenter
	service   deployedProductsService
}

//go:generate counterfeiter -o ./fakes/deployed_products_service.go --fake-name DeployedProductsService . deployedProductsService
type deployedProductsService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewDeployedProducts(presenter presenters.Presenter, service deployedProductsService) DeployedProducts {
	return DeployedProducts{
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

	dp.presenter.PresentDeployedProducts(deployedProducts)

	return nil
}

func (dp DeployedProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all deployed products.",
		ShortDescription: "lists deployed products",
	}
}
