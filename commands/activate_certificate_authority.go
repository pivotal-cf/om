package commands

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"github.com/pivotal-cf/om/api"
)

type ActivateCertificateAuthority struct {
	service activateCertificateAuthorityService
	logger  logger
	Options struct {
		Id string `long:"id" required:"true" description:"certificate authority id"`
	}
}

//counterfeiter:generate -o ./fakes/activate_certificate_authority_service.go --fake-name ActivateCertificateAuthorityService . activateCertificateAuthorityService
type activateCertificateAuthorityService interface {
	ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput) error
}

func NewActivateCertificateAuthority(service activateCertificateAuthorityService, logger logger) *ActivateCertificateAuthority {
	return &ActivateCertificateAuthority{service: service, logger: logger}
}

func (a ActivateCertificateAuthority) Execute(args []string) error {
	err := a.service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' activated\n", a.Options.Id)

	return nil
}
