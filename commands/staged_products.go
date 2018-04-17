package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type StagedProducts struct {
	presenter presenters.Presenter
	service   stagedProductsService
}

//go:generate counterfeiter -o ./fakes/staged_products_service.go --fake-name StagedProductsService . stagedProductsService
type stagedProductsService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewStagedProducts(presenter presenters.Presenter, service stagedProductsService) StagedProducts {
	return StagedProducts{
		presenter: presenter,
		service:   service,
	}
}

func (sp StagedProducts) Execute(args []string) error {
	diagnosticReport, err := sp.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve staged products %s", err)
	}

	stagedProducts := diagnosticReport.StagedProducts

	sp.presenter.PresentStagedProducts(stagedProducts)
	return nil
}

func (sp StagedProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all staged products.",
		ShortDescription: "lists staged products",
	}
}
