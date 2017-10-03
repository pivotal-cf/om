package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type DeleteCertificateAuthority struct {
	service certificateAuthorityDeleter
	logger  logger
	Options struct {
		Id string `long:"id"  description:"certificate authority id"`
	}
}

//go:generate counterfeiter -o ./fakes/certificate_authority_deleter.go --fake-name CertificateAuthorityDeleter . certificateAuthorityDeleter
type certificateAuthorityDeleter interface {
	Delete(api.DeleteCertificateAuthorityInput) error
}

func NewDeleteCertificateAuthority(service certificateAuthorityDeleter, logger logger) DeleteCertificateAuthority {
	return DeleteCertificateAuthority{service: service, logger: logger}
}

func (a DeleteCertificateAuthority) Execute(args []string) error {
	_, err := flags.Parse(&a.Options, args)

	if err != nil {
		return fmt.Errorf("could not parse delete-certificate-authority flags: %s", err)
	}

	if a.Options.Id == "" {
		return errors.New("error: id is missing. Please see usage for more information.")
	}

	err = a.service.Delete(api.DeleteCertificateAuthorityInput{
		GUID: a.Options.Id,
	})

	if err != nil {
		return err
	}

	a.logger.Printf("Certificate authority '%s' deleted\n", a.Options.Id)

	return nil
}

func (a DeleteCertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command deletes an existing certificate authority on the Ops Manager",
		ShortDescription: "deletes a certificate authority on the Ops Manager",
		Flags:            a.Options,
	}
}
