package commands

import (
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

//counterfeiter:generate -o ./fakes/generate_certificate_authority_service.go --fake-name GenerateCertificateAuthorityService . generateCertificateAuthorityService
type generateCertificateAuthorityService interface {
	GenerateCertificateAuthority() (api.CA, error)
}

func NewGenerateCertificateAuthority(service generateCertificateAuthorityService, presenter presenters.FormattedPresenter) *GenerateCertificateAuthority {
	return &GenerateCertificateAuthority{service: service, presenter: presenter}
}

func (g GenerateCertificateAuthority) Execute(args []string) error {
	certificateAuthority, err := g.service.GenerateCertificateAuthority()
	if err != nil {
		return err
	}

	g.presenter.SetFormat(g.Options.Format)
	g.presenter.PresentCertificateAuthority(certificateAuthority)

	return nil
}
