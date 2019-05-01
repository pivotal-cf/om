package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type DiagnosticReport struct {
	presenter presenters.FormattedPresenter
	service   diagnosticReportService
	Options   struct {
	}
}

//go:generate counterfeiter -o ./fakes/diagnostic_report_service.go --fake-name DiagnosticReportService . diagnosticReportService
type diagnosticReportService interface {
	GetDiagnosticReport() (api.DiagnosticReport, error)
}

func NewDiagnosticReport(presenter presenters.FormattedPresenter, service diagnosticReportService) DiagnosticReport {
	return DiagnosticReport{
		presenter: presenter,
		service:   service,
	}
}

func (dr DiagnosticReport) Execute(args []string) error {
	if _, err := jhanda.Parse(&dr.Options, args); err != nil {
		return fmt.Errorf("could not parse diagnostic-report flags: %s", err)
	}

	diagnosticReport, err := dr.service.GetDiagnosticReport()
	if err != nil {
		return fmt.Errorf("failed to retrieve diagnostic-report %s", err)
	}

	dr.presenter.SetFormat("json")
	dr.presenter.PresentDiagnosticReport(diagnosticReport)

	return nil
}

func (dr DiagnosticReport) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "retrieve a diagnostic report with general information about the state of your Ops Manager.",
		ShortDescription: "reports current state of your Ops Manager",
		Flags:            dr.Options,
	}
}
