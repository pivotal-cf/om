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
	GenerateCertificateAuthority() (api.GenerateCAResponse, error)
}

func NewGenerateCertificateAuthority(service generateCertificateAuthorityService, presenter presenters.FormattedPresenter) *GenerateCertificateAuthority {
	return &GenerateCertificateAuthority{service: service, presenter: presenter}
}

func (g GenerateCertificateAuthority) Execute(args []string) error {
	caResp, err := g.service.GenerateCertificateAuthority()
	if err != nil {
		return err
	}

	g.presenter.SetFormat(g.Options.Format)
	if caResp.Warnings != nil {
		g.presenter.PresentGenerateCAResponse(caResp)
	} else {
		g.presenter.PresentCertificateAuthority(caResp.CA)
	}

	return nil
}
