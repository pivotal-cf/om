package commands

import (
	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthorities struct {
	cas       certificateAuthorityLister
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/certificate_authority_lister.go --fake-name CertificateAuthorityLister . certificateAuthorityLister
type certificateAuthorityLister interface {
	List() (api.CertificateAuthoritiesOutput, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthorityLister, presenter presenters.Presenter) CertificateAuthorities {
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

func (c CertificateAuthorities) Usage() commands.Usage {
	return commands.Usage{
		Description:      "lists certificates managed by Ops Manager",
		ShortDescription: "lists certificates managed by Ops Manager",
	}
}
