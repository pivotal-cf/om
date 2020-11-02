package commands

import (
	"fmt"

	"github.com/pivotal-cf/om/presenters"
)

type StagedProducts struct {
	presenter presenters.FormattedPresenter
	service   diagnosticReportService
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewStagedProducts(presenter presenters.FormattedPresenter, service diagnosticReportService) *StagedProducts {
	return &StagedProducts{
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

	sp.presenter.SetFormat(sp.Options.Format)
	sp.presenter.PresentStagedProducts(stagedProducts)

	return nil
}
