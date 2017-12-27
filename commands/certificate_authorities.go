package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthorities struct {
	cas       certificateAuthoritiesService
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/certificate_authorities_service.go --fake-name CertificateAuthoritiesService . certificateAuthoritiesService

type certificateAuthoritiesService interface {
	List() (api.CertificateAuthoritiesOutput, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.Presenter) CertificateAuthorities {
	return CertificateAuthorities{
		cas:       certificateAuthoritiesService,
		presenter: presenter,
	}
}

func (c CertificateAuthorities) Execute(_ []string) error {
	casOutput, err := c.cas.List()
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
