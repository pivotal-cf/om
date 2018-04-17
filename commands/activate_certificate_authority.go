package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ActivateCertificateAuthority struct {
	service activateCertificateAuthorityService
	logger  logger
	Options struct {
		Id string `long:"id" required:"true" description:"certificate authority id"`
	}
}

//go:generate counterfeiter -o ./fakes/activate_certificate_authority_service.go --fake-name ActivateCertificateAuthorityService . activateCertificateAuthorityService
type activateCertificateAuthorityService interface {
	ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput) error
}

func NewActivateCertificateAuthority(service activateCertificateAuthorityService, logger logger) ActivateCertificateAuthority {
	return ActivateCertificateAuthority{service: service, logger: logger}
}

func (a ActivateCertificateAuthority) Execute(args []string) error {
	if _, err := jhanda.Parse(&a.Options, args); err != nil {
		return fmt.Errorf("could not parse activate-certificate-authority flags: %s", err)
	}

	err := a.service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' activated\n", a.Options.Id)

	return nil
}

func (a ActivateCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command activates an existing certificate authority on the Ops Manager",
		ShortDescription: "activates a certificate authority on the Ops Manager",
		Flags:            a.Options,
	}
}
