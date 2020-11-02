package commands

import (
	"github.com/pivotal-cf/om/api"
)

type DeleteCertificateAuthority struct {
	service deleteCertificateAuthorityService
	logger  logger
	Options struct {
		Id string `long:"id" required:"true" description:"certificate authority id"`
	}
}

//counterfeiter:generate -o ./fakes/delete_certificate_authority_service.go --fake-name DeleteCertificateAuthorityService . deleteCertificateAuthorityService
type deleteCertificateAuthorityService interface {
	DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput) error
}

func NewDeleteCertificateAuthority(service deleteCertificateAuthorityService, logger logger) *DeleteCertificateAuthority {
	return &DeleteCertificateAuthority{service: service, logger: logger}
}

func (a DeleteCertificateAuthority) Execute(args []string) error {
	err := a.service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' deleted\n", a.Options.Id)

	return nil
}
