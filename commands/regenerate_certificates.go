package commands

type RegenerateCertificates struct {
	service regenerateCertificatesService
	logger  logger
}

//counterfeiter:generate -o ./fakes/regenerate_certificates_service.go --fake-name RegenerateCertificatesService . regenerateCertificatesService
type regenerateCertificatesService interface {
	RegenerateCertificates() error
}

func NewRegenerateCertificates(service regenerateCertificatesService, logger logger) *RegenerateCertificates {
	return &RegenerateCertificates{service: service, logger: logger}
}

func (r RegenerateCertificates) Execute(_ []string) error {
	err := r.service.RegenerateCertificates()
	if err != nil {
		return err
	}

	r.logger.Printf("Certificates regenerated.\n")

	return nil
}
