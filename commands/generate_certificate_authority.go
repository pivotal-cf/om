package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type GenerateCertificateAuthority struct {
	service   generateCertificateAuthorityService
	presenter presenters.FormattedPresenter
	Options   struct {
		Format string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//go:generate counterfeiter -o ./fakes/generate_certificate_authority_service.go --fake-name GenerateCertificateAuthorityService . generateCertificateAuthorityService
type generateCertificateAuthorityService interface {
	GenerateCertificateAuthority() (api.CA, error)
}

func NewGenerateCertificateAuthority(service generateCertificateAuthorityService, presenter presenters.FormattedPresenter) GenerateCertificateAuthority {
	return GenerateCertificateAuthority{service: service, presenter: presenter}
}

func (g GenerateCertificateAuthority) Execute(args []string) error {
	if _, err := jhanda.Parse(&g.Options, args); err != nil {
		return fmt.Errorf("could not parse generate-certificate-authority flags: %s", err)
	}

	certificateAuthority, err := g.service.GenerateCertificateAuthority()
	if err != nil {
		return err
	}

	g.presenter.SetFormat(g.Options.Format)
	g.presenter.PresentCertificateAuthority(certificateAuthority)

	return nil
}

func (g GenerateCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command generates a certificate authority on the Ops Manager",
		ShortDescription: "generates a certificate authority on the Opsman",
		Flags:            g.Options,
	}
}
