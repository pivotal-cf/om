package commands

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
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
		caList, err := a.service.ListCertificateAuthorities()

		if err != nil {
			return err
		}

		latestCA := getLatestCertificateAuthority(caList.CAs)

		if latestCA.Active {
			a.logger.Printf("No newer certificate authority available to activate\n")
			return nil
		}
		guid = latestCA.GUID
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

func getLatestCertificateAuthority(caList []api.CA) api.CA {
	var newestCA api.CA
	var newestCACreationTime time.Time
	for _, ca := range caList {
		tmp, _ := time.Parse(time.RFC3339, ca.CreatedOn)
		if tmp.After(newestCACreationTime) {
			newestCA = ca
			newestCACreationTime = tmp
		}
	}
	return newestCA
}
