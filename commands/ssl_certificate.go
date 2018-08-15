package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
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

//go:generate counterfeiter -o ./fakes/ssl_certificate_service.go --fake-name SSLCertificateService . sslCertificateService

type sslCertificateService interface {
	GetSSLCertificate() (api.SSLCertificateOutput, error)
}

func NewSSLCertificate(service sslCertificateService, presenter presenters.FormattedPresenter) SSLCertificate {
	return SSLCertificate{
		service:   service,
		presenter: presenter,
	}
}

func (c SSLCertificate) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse ssl-certificate flags: %s", err)
	}

	certOutput, err := c.service.GetSSLCertificate()
	if err != nil {
		return err
	}

	c.presenter.SetFormat(c.Options.Format)
	c.presenter.PresentSSLCertificate(certOutput.Certificate)

	return nil
}

func (c SSLCertificate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command gets certificate applied to Ops Manager",
		ShortDescription: "gets certificate applied to Ops Manager",
		Flags:            c.Options,
	}
}
