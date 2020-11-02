package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type SSLCertificate struct {
	service   sslCertificateService
	presenter presenters.FormattedPresenter
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//counterfeiter:generate -o ./fakes/ssl_certificate_service.go --fake-name SSLCertificateService . sslCertificateService

type sslCertificateService interface {
	GetSSLCertificate() (api.SSLCertificateOutput, error)
}

func NewSSLCertificate(service sslCertificateService, presenter presenters.FormattedPresenter) *SSLCertificate {
	return &SSLCertificate{
		service:   service,
		presenter: presenter,
	}
}

func (c SSLCertificate) Execute(args []string) error {
	certOutput, err := c.service.GetSSLCertificate()
	if err != nil {
		return err
	}

	c.presenter.SetFormat(c.Options.Format)
	c.presenter.PresentSSLCertificate(certOutput.Certificate)

	return nil
}
