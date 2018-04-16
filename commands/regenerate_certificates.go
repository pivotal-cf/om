package commands

import "github.com/pivotal-cf/jhanda"

type RegenerateCertificates struct {
	service certificateRegenerator
	logger  logger
}

//go:generate counterfeiter -o ./fakes/certificate_regenerator.go --fake-name CertificateRegenerator . certificateRegenerator
type certificateRegenerator interface {
	Regenerate() error
}

func NewRegenerateCertificates(service certificateRegenerator, logger logger) RegenerateCertificates {
	return RegenerateCertificates{service: service, logger: logger}
}

func (r RegenerateCertificates) Execute(_ []string) error {
	err := r.service.Regenerate()
	if err != nil {
		return err
	}

	r.logger.Printf("Certificates regenerated.\n")

	return nil
}

func (r RegenerateCertificates) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes all non-configurable certificates in Ops Manager",
		ShortDescription: "deletes all non-configurable certificates in Ops Manager",
	}
}
