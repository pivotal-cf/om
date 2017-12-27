package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type GenerateCertificateAuthority struct {
	service   certificateAuthorityGenerator
	presenter presenters.Presenter
}

//go:generate counterfeiter -o ./fakes/certificate_authority_generator.go --fake-name CertificateAuthorityGenerator . certificateAuthorityGenerator
type certificateAuthorityGenerator interface {
	Generate() (api.CA, error)
}

func NewGenerateCertificateAuthority(service certificateAuthorityGenerator, presenter presenters.Presenter) GenerateCertificateAuthority {
	return GenerateCertificateAuthority{service: service, presenter: presenter}
}

func (g GenerateCertificateAuthority) Execute(_ []string) error {
	certificateAuthority, err := g.service.Generate()
	if err != nil {
		return err
	}

	g.presenter.PresentCertificateAuthority(certificateAuthority)

	return nil
}

func (g GenerateCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command generates a certificate authority on the Ops Manager",
		ShortDescription: "generates a certificate authority on the Opsman",
	}
}
