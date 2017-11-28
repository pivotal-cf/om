package commands

import (
	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
)

type GenerateCertificateAuthority struct {
	service   certificateAuthorityGenerator
	presenter Presenter
}

//go:generate counterfeiter -o ./fakes/certificate_authority_generator.go --fake-name CertificateAuthorityGenerator . certificateAuthorityGenerator
type certificateAuthorityGenerator interface {
	Generate() (api.CA, error)
}

func NewGenerateCertificateAuthority(service certificateAuthorityGenerator, presenter Presenter) GenerateCertificateAuthority {
	return GenerateCertificateAuthority{service: service, presenter: presenter}
}

func (g GenerateCertificateAuthority) Execute(_ []string) error {
	certificateAuthority, err := g.service.Generate()
	if err != nil {
		return err
	}

	g.presenter.PresentGeneratedCertificateAuthority(certificateAuthority)

	return nil
}

func (g GenerateCertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command generates a certificate authority on the Ops Manager",
		ShortDescription: "generates a certificate authority on the Opsman",
	}
}
