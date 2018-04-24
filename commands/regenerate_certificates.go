package commands

import "github.com/pivotal-cf/jhanda"

type RegenerateCertificates struct {
	service regenerateCertificatesService
	logger  logger
}

//go:generate counterfeiter -o ./fakes/regenerate_certificates_service.go --fake-name RegenerateCertificatesService . regenerateCertificatesService
type regenerateCertificatesService interface {
	RegenerateCertificates() error
}

func NewRegenerateCertificates(service regenerateCertificatesService, logger logger) RegenerateCertificates {
	return RegenerateCertificates{service: service, logger: logger}
}

func (r RegenerateCertificates) Execute(_ []string) error {
	err := r.service.RegenerateCertificates()
	if err != nil {
		return err
	}

	r.logger.Printf("Certificates regenerated.\n")

	return nil
}

func (r RegenerateCertificates) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes",
		ShortDescription: "deletes all non-configurable certificates in Ops Manager so they will automatically be regenerated on the next apply-changes",
	}
}
