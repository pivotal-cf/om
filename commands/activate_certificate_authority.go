package commands

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"time"

	"github.com/pivotal-cf/om/api"
)

type ActivateCertificateAuthority struct {
	service activateCertificateAuthorityService
	logger  logger
	Options struct {
		Id string `long:"id" required:"false" description:"certificate authority id"`
	}
}

//counterfeiter:generate -o ./fakes/activate_certificate_authority_service.go --fake-name ActivateCertificateAuthorityService . activateCertificateAuthorityService
type activateCertificateAuthorityService interface {
	ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput) error
	ListCertificateAuthorities() (api.CertificateAuthoritiesOutput, error)
}

func NewActivateCertificateAuthority(service activateCertificateAuthorityService, logger logger) *ActivateCertificateAuthority {
	return &ActivateCertificateAuthority{service: service, logger: logger}
}

func (a ActivateCertificateAuthority) Execute(args []string) error {
	guid := a.Options.Id
	if a.Options.Id == "" {
		caList, _ := a.service.ListCertificateAuthorities()
		var activeCA api.CA
		var inactiveCA api.CA

		for _, ca := range caList.CAs {
			if ca.Active {
				activeCA = ca
			} else {
				inactiveCA = ca
			}
		}

		if inactiveCA.GUID == "" {
			return fmt.Errorf("no inactive certificate authorities to activate")
		}

		inactiveCreationTime, _ := time.Parse(time.RFC3339, inactiveCA.CreatedOn)
		fmt.Printf("%+v\n", inactiveCreationTime)
		activeCreationTime, _ := time.Parse(time.RFC3339, activeCA.CreatedOn)
		fmt.Printf("%+v\n", activeCreationTime)

		if activeCreationTime.After(inactiveCreationTime) {
			a.logger.Printf("No newer certificate authority available to activate\n")
			return nil
		}

		guid = inactiveCA.GUID
	}

	err := a.service.ActivateCertificateAuthority(api.ActivateCertificateAuthorityInput{
		GUID: guid,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' activated\n", a.Options.Id)

	return nil
}
