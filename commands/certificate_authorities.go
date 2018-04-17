package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthorities struct {
	service   certificateAuthoritiesService
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/certificate_authorities_service.go --fake-name CertificateAuthoritiesService . certificateAuthoritiesService

type certificateAuthoritiesService interface {
	ListCertificateAuthorities() (api.CertificateAuthoritiesOutput, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.Presenter) CertificateAuthorities {
	return CertificateAuthorities{
		service:   certificateAuthoritiesService,
		presenter: presenter,
	}
}

func (c CertificateAuthorities) Execute(_ []string) error {
	casOutput, err := c.service.ListCertificateAuthorities()
	if err != nil {
		return err
	}

	c.presenter.PresentCertificateAuthorities(casOutput.CAs)

	return nil
}

func (c CertificateAuthorities) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "lists certificates managed by Ops Manager",
		ShortDescription: "lists certificates managed by Ops Manager",
	}
}
