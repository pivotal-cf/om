package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type CreateCertificateAuthority struct {
	service   certificateAuthorityCreator
	presenter Presenter
	Options   struct {
		CertPem    string `long:"certificate-pem"  description:"certificate"`
		PrivateKey string `long:"private-key-pem"  description:"private key"`
	}
}

//go:generate counterfeiter -o ./fakes/certificate_authority_creator.go --fake-name CertificateAuthorityCreator . certificateAuthorityCreator
type certificateAuthorityCreator interface {
	Create(api.CertificateAuthorityInput) (api.CA, error)
}

func NewCreateCertificateAuthority(service certificateAuthorityCreator, presenter Presenter) CreateCertificateAuthority {
	return CreateCertificateAuthority{service: service, presenter: presenter}
}

func (c CreateCertificateAuthority) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse create-certificate-authority flags: %s", err)
	}

	if c.Options.CertPem == "" {
		return errors.New("error: certificate-pem is missing. Please see usage for more information.")
	}

	if c.Options.PrivateKey == "" {
		return errors.New("error: private-key-pem is missing. Please see usage for more information.")
	}

	ca, err := c.service.Create(api.CertificateAuthorityInput{
		CertPem:       c.Options.CertPem,
		PrivateKeyPem: c.Options.PrivateKey,
	})
	if err != nil {
		return err
	}

	c.presenter.PresentCertificateAuthority(ca)

	return nil
}

func (c CreateCertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command creates a certificate authority on the Ops Manager with the given cert and key",
		ShortDescription: "creates a certificate authority on the Ops Manager",
		Flags:            c.Options,
	}
}
