package commands

import (
	"fmt"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DiagnosticReport struct {
	presenter presenters.FormattedPresenter
	service   diagnosticReportService
	Options   struct {
	}
}

//counterfeiter:generate -o ./fakes/diagnostic_report_service.go --fake-name DiagnosticReportService . diagnosticReportService
type diagnosticReportService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewDiagnosticReport(presenter presenters.FormattedPresenter, service diagnosticReportService) *DiagnosticReport {
	return &DiagnosticReport{
		presenter: presenter,
		service:   service,
	}
}

func (dr DiagnosticReport) Execute(args []string) error {
	diagnosticReport, err := dr.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve diagnostic-report %s", err)
	}

	dr.presenter.SetFormat("json")
	dr.presenter.PresentDiagnosticReport(diagnosticReport)

	return nil
}
