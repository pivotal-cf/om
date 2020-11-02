package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthorities struct {
	service   certificateAuthoritiesService
	presenter presenters.FormattedPresenter
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//counterfeiter:generate -o ./fakes/certificate_authorities_service.go --fake-name CertificateAuthoritiesService . certificateAuthoritiesService

type certificateAuthoritiesService interface {
	ListCertificateAuthorities() (api.CertificateAuthoritiesOutput, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.FormattedPresenter) *CertificateAuthorities {
	return &CertificateAuthorities{
		service:   certificateAuthoritiesService,
		presenter: presenter,
	}
}

func (c CertificateAuthorities) Execute(args []string) error {
	casOutput, err := c.service.ListCertificateAuthorities()
	if err != nil {
		return err
	}

	c.presenter.SetFormat(c.Options.Format)
	c.presenter.PresentCertificateAuthorities(casOutput.CAs)

	return nil
}
