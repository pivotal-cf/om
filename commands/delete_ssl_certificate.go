package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type DeleteSSLCertificate struct {
	service deleteSSLCertificateService
	logger  logger
	Options struct{}
}

//go:generate counterfeiter -o ./fakes/delete_ssl_certificate_service.go --fake-name DeleteSSLCertificateService . deleteSSLCertificateService

type deleteSSLCertificateService interface {
	DeleteSSLCertificate() error
}

func NewDeleteSSLCertificate(service deleteSSLCertificateService, logger logger) DeleteSSLCertificate {
	return DeleteSSLCertificate{
		service: service,
		logger:  logger,
	}
}

func (c DeleteSSLCertificate) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse delete-ssl-certificate flags: %s", err)
	}

	err := c.service.DeleteSSLCertificate()
	if err != nil {
		return err
	}

	c.logger.Printf("Successfully deleted custom SSL Certificate and reverted to the provided self-signed SSL certificate.\n")
	c.logger.Printf("Please allow about 1 min for the new certificate to take effect.\n")

	return nil
}

func (c DeleteSSLCertificate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes a custom certificate applied to Ops Manager and reverts to the auto-generated cert",
		ShortDescription: "deletes certificate applied to Ops Manager",
		Flags:            c.Options,
	}
}
