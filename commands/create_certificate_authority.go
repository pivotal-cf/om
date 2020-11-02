package commands

import (
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CreateCertificateAuthority struct {
	service   createCertificateAuthorityService
	presenter presenters.FormattedPresenter
	Options   struct {
		CertPem    string `long:"certificate-pem" required:"true" description:"certificate"`
		PrivateKey string `long:"private-key-pem" required:"true" description:"private key"`
		Format     string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

//counterfeiter:generate -o ./fakes/create_certificate_authority_service.go --fake-name CreateCertificateAuthorityService . createCertificateAuthorityService
type createCertificateAuthorityService interface {
	CreateCertificateAuthority(api.CertificateAuthorityInput) (api.CA, error)
}

func NewCreateCertificateAuthority(service createCertificateAuthorityService, presenter presenters.FormattedPresenter) *CreateCertificateAuthority {
	return &CreateCertificateAuthority{service: service, presenter: presenter}
}

func (c CreateCertificateAuthority) Execute(args []string) error {
	ca, err := c.service.CreateCertificateAuthority(api.CertificateAuthorityInput{
		CertPem:       c.Options.CertPem,
		PrivateKeyPem: c.Options.PrivateKey,
	})
	if err != nil {
		return err
	}

	c.presenter.SetFormat(c.Options.Format)
	c.presenter.PresentCertificateAuthority(ca)

	return nil
}
