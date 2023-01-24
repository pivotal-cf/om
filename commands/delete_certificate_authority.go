package commands

import (
	"errors"

	"github.com/pivotal-cf/om/api"
)

type DeleteCertificateAuthority struct {
	service deleteCertificateAuthorityService
	logger  logger
	Options struct {
		Id          string `long:"id" required:"false" description:"certificate authority id"`
		AllInactive bool   `long:"all-inactive" required:"false" description:"delete all inactive certificate authorities"`
	}
}

//counterfeiter:generate -o ./fakes/delete_certificate_authority_service.go --fake-name DeleteCertificateAuthorityService . deleteCertificateAuthorityService
type deleteCertificateAuthorityService interface {
	DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput) error
	ListCertificateAuthorities() (api.CertificateAuthoritiesOutput, error)
}

func NewDeleteCertificateAuthority(service deleteCertificateAuthorityService, logger logger) *DeleteCertificateAuthority {
	return &DeleteCertificateAuthority{service: service, logger: logger}
}

func (a DeleteCertificateAuthority) Execute(args []string) error {
	var caGuid string
	var err error

	switch {
	case a.Options.AllInactive:
		if caGuid, err = a.getInactiveCA(); err != nil {
			return err
		}
	case a.Options.Id != "":
		caGuid = a.Options.Id
	default:
		return errors.New("--id or --all-inactive must be provided")
	}

	err = a.service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
		GUID: caGuid,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' deleted\n", caGuid)

	return nil
}

func (a DeleteCertificateAuthority) getInactiveCA() (string, error) {
	caList, err := a.service.ListCertificateAuthorities()
	if err != nil {
		return "", err
	}

	for _, ca := range caList.CAs {
		if !ca.Active {
			return ca.GUID, nil
		}
	}

	return "", errors.New("no inactive CAs found")
}
