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
	var caGuids []string
	var err error

	switch {
	case a.Options.AllInactive:
		if caGuids, err = a.getInactiveCAs(); err != nil {
			return err
		}
	case a.Options.Id != "":
		caGuids = append(caGuids, a.Options.Id)
	default:
		return errors.New("--id or --all-inactive must be provided")
	}

	for _, caGuid := range caGuids {
		err = a.service.DeleteCertificateAuthority(api.DeleteCertificateAuthorityInput{
			GUID: caGuid,
		})
		if err != nil {
			return err
		}

		a.logger.Printf("Certificate authority '%s' deleted\n", caGuid)
	}

	return nil
}

func (a DeleteCertificateAuthority) getInactiveCAs() ([]string, error) {
	caList, err := a.service.ListCertificateAuthorities()
	out := []string{}
	if err != nil {
		return nil, err
	}

	for _, ca := range caList.CAs {
		if !ca.Active {
			out = append(out, ca.GUID)
		}
	}

	if len(out) != 0 {
		return out, nil
	}

	return nil, errors.New("no inactive CAs found")
}
