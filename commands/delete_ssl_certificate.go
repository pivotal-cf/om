package commands

type DeleteSSLCertificate struct {
	service deleteSSLCertificateService
	logger  logger
	Options struct{}
}

//counterfeiter:generate -o ./fakes/delete_ssl_certificate_service.go --fake-name DeleteSSLCertificateService . deleteSSLCertificateService

type deleteSSLCertificateService interface {
	DeleteSSLCertificate() error
}

func NewDeleteSSLCertificate(service deleteSSLCertificateService, logger logger) *DeleteSSLCertificate {
	return &DeleteSSLCertificate{
		service: service,
		logger:  logger,
	}
}

func (c DeleteSSLCertificate) Execute(args []string) error {
	err := c.service.DeleteSSLCertificate()
	if err != nil {
		return err
	}

	c.logger.Printf("Successfully deleted custom SSL Certificate and reverted to the provided self-signed SSL certificate.\n")
	c.logger.Printf("Please allow about 1 min for the new certificate to take effect.\n")

	return nil
}
