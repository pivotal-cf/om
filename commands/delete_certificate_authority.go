package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeleteCertificateAuthority struct {
	service deleteCertificateAuthorityService
	logger  logger
	Options struct {
		Id string `long:"id" required:"true" description:"certificate authority id"`
	}
}

//go:generate counterfeiter -o ./fakes/delete_certificate_authority_service.go --fake-name DeleteCertificateAuthorityService . deleteCertificateAuthorityService
type deleteCertificateAuthorityService interface {
	DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput) error
}

func NewDeleteCertificateAuthority(service deleteCertificateAuthorityService, logger logger) DeleteCertificateAuthority {
	return DeleteCertificateAuthority{service: service, logger: logger}
}

func (a DeleteCertificateAuthority) Execute(args []string) error {
	if _, err := jhanda.Parse(&a.Options, args); err != nil {
		return fmt.Errorf("could not parse delete-certificate-authority flags: %s", err)
	}

	err := a.service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' deleted\n", a.Options.Id)

	return nil
}

func (a DeleteCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes an existing certificate authority on the Ops Manager",
		ShortDescription: "deletes a certificate authority on the Ops Manager",
		Flags:            a.Options,
	}
}
