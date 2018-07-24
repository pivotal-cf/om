package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
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

//go:generate counterfeiter -o ./fakes/create_certificate_authority_service.go --fake-name CreateCertificateAuthorityService . createCertificateAuthorityService
type createCertificateAuthorityService interface {
	CreateCertificateAuthority(api.CertificateAuthorityInput) (api.CA, error)
}

func NewCreateCertificateAuthority(service createCertificateAuthorityService, presenter presenters.FormattedPresenter) CreateCertificateAuthority {
	return CreateCertificateAuthority{service: service, presenter: presenter}
}

func (c CreateCertificateAuthority) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse create-certificate-authority flags: %s", err)
	}

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

func (c CreateCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command creates a certificate authority on the Ops Manager with the given cert and key",
		ShortDescription: "creates a certificate authority on the Ops Manager",
		Flags:            c.Options,
	}
}
