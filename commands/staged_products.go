package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type StagedProducts struct {
	presenter presenters.FormattedPresenter
	service   stagedProductsService
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/staged_products_service.go --fake-name StagedProductsService . stagedProductsService
type stagedProductsService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewStagedProducts(presenter presenters.FormattedPresenter, service stagedProductsService) StagedProducts {
	return StagedProducts{
		presenter: presenter,
		service:   service,
	}
}

func (sp StagedProducts) Execute(args []string) error {
	if _, err := jhanda.Parse(&sp.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-products flags: %s", err)
	}

	diagnosticReport, err := sp.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve staged products %s", err)
	}

	stagedProducts := diagnosticReport.StagedProducts

	sp.presenter.SetFormat(sp.Options.Format)
	sp.presenter.PresentStagedProducts(stagedProducts)

	return nil
}

func (sp StagedProducts) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command lists all staged products.",
		ShortDescription: "lists staged products",
		Flags:            sp.Options,
	}
}
