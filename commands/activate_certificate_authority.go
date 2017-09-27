package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type ActivateCertificateAuthority struct {
	service certificateAuthorityActivator
	logger  logger
	Options struct {
		Id string `long:"id"  description:"certificate authority id"`
	}
}

//go:generate counterfeiter -o ./fakes/certificate_authority_activator.go --fake-name CertificateAuthorityActivator . certificateAuthorityActivator
type certificateAuthorityActivator interface {
	Activate(api.ActivateCertificateAuthorityInput) error
}

func NewActivateCertificateAuthority(service certificateAuthorityActivator, logger logger) ActivateCertificateAuthority {
	return ActivateCertificateAuthority{service: service, logger: logger}
}

func (a ActivateCertificateAuthority) Execute(args []string) error {
	_, err := flags.Parse(&a.Options, args)

	if err != nil {
		return fmt.Errorf("could not parse activate-certificate-authority flags: %s", err)
	}

	if a.Options.Id == "" {
		return errors.New("error: id is missing. Please see usage for more information.")
	}

	err = a.service.Activate(api.ActivateCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' activated\n", a.Options.Id)

	return nil
}

func (a ActivateCertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command activates an existing certificate authority on the Ops Manager",
		ShortDescription: "activates a certificate authority on the Ops Manager",
		Flags:            a.Options,
	}
}
