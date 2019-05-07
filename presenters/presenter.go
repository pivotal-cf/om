package presenters

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/models"
)

//go:generate counterfeiter -o fakes/presenter.go --fake-name Presenter . Presenter

type Presenter interface {
	PresentAvailableProducts([]models.Product)
	PresentCertificateAuthorities([]api.CA)
	PresentCertificateAuthority(api.CA)
	PresentSSLCertificate(api.SSLCertificate)
	PresentCredentialReferences([]string)
	PresentCredentials(map[string]string)
	PresentDeployedProducts([]api.DiagnosticProduct)
	PresentErrands([]models.Errand)
	PresentInstallations([]models.Installation)
	PresentPendingChanges(api.PendingChangesOutput)
	PresentStagedProducts([]api.DiagnosticProduct)
	PresentDiagnosticReport(api.DiagnosticReport)
}

//go:generate counterfeiter -o fakes/formatted_presenter.go --fake-name FormattedPresenter . FormattedPresenter

type FormattedPresenter interface {
	Presenter
	SetFormat(string)
}

type MultiPresenter struct {
	tablePresenter Presenter
	jsonPresenter  Presenter
	format         string
}

func NewPresenter(tablePresenter Presenter, jsonPresenter Presenter) *MultiPresenter {
	return &MultiPresenter{
		tablePresenter: tablePresenter,
		jsonPresenter:  jsonPresenter,
		format:         "table",
	}
}

func (p *MultiPresenter) SetFormat(format string) {
	p.format = format
}

func (p *MultiPresenter) PresentAvailableProducts(products []models.Product) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentAvailableProducts(products)
	default:
		p.tablePresenter.PresentAvailableProducts(products)
	}
}

func (p *MultiPresenter) PresentCertificateAuthorities(cas []api.CA) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentCertificateAuthorities(cas)
	default:
		p.tablePresenter.PresentCertificateAuthorities(cas)
	}
}

func (p *MultiPresenter) PresentCertificateAuthority(ca api.CA) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentCertificateAuthority(ca)
	default:
		p.tablePresenter.PresentCertificateAuthority(ca)
	}
}

func (p *MultiPresenter) PresentSSLCertificate(cert api.SSLCertificate) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentSSLCertificate(cert)
	default:
		p.tablePresenter.PresentSSLCertificate(cert)
	}
}

func (p *MultiPresenter) PresentCredentialReferences(ref []string) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentCredentialReferences(ref)
	default:
		p.tablePresenter.PresentCredentialReferences(ref)
	}
}

func (p *MultiPresenter) PresentCredentials(creds map[string]string) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentCredentials(creds)
	default:
		p.tablePresenter.PresentCredentials(creds)
	}
}

func (p *MultiPresenter) PresentDeployedProducts(products []api.DiagnosticProduct) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentDeployedProducts(products)
	default:
		p.tablePresenter.PresentDeployedProducts(products)
	}
}

func (p *MultiPresenter) PresentErrands(errands []models.Errand) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentErrands(errands)
	default:
		p.tablePresenter.PresentErrands(errands)
	}
}

func (p *MultiPresenter) PresentInstallations(i []models.Installation) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentInstallations(i)
	default:
		p.tablePresenter.PresentInstallations(i)
	}
}
func (p *MultiPresenter) PresentPendingChanges(c api.PendingChangesOutput) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentPendingChanges(c)
	default:
		p.tablePresenter.PresentPendingChanges(c)
	}
}

func (p *MultiPresenter) PresentStagedProducts(products []api.DiagnosticProduct) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentStagedProducts(products)
	default:
		p.tablePresenter.PresentStagedProducts(products)
	}
}

func (p *MultiPresenter) PresentDiagnosticReport(report api.DiagnosticReport) {
	switch p.format {
	case "json":
		p.jsonPresenter.PresentDiagnosticReport(report)
	default:
		p.tablePresenter.PresentDiagnosticReport(report)
	}
}
