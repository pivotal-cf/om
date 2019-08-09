package commands

import (
	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
)

//go:generate counterfeiter -o ./fakes/expiring_certs_service.go --fake-name ExpiringCertsService . expiringCertsService
type expiringCertsService interface {
	ListExpiringCertificates(string) ([]api.ExpiringCertificate, error)
}

type ExpiringCerts struct {
	logger logger
	api    expiringCertsService
}

func NewExpiringCertificates(service expiringCertsService, logger logger) *ExpiringCerts {
	return &ExpiringCerts{
		api:    service,
		logger: logger,
	}
}

func (e *ExpiringCerts) Execute(args []string) error {
	defaultDuration := "3m"
	expiringCerts, _ := e.api.ListExpiringCertificates(defaultDuration)

	e.logger.Println("Getting expiring certificates...")
	if len(expiringCerts) > 0 {
		e.logger.Println(color.RedString("[X] Ops Manager"))
		e.logger.Println(color.RedString("[X] Credhub"))
	} else {
		e.logger.Println(color.GreenString("[✓] Ops Manager"))
		e.logger.Println(color.GreenString("[✓] Credhub"))
	}

	return nil
}
