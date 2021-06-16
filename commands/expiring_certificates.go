package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pivotal-cf/om/api"
)

//counterfeiter:generate -o ./fakes/expiring_certs_service.go --fake-name ExpiringCertsService . expiringCertsService
type expiringCertsService interface {
	ListExpiringCertificates(string) ([]api.ExpiringCertificate, error)
}

type ExpiringCerts struct {
	logger  logger
	api     expiringCertsService
	Options struct {
		ExpiresWithin string `long:"expires-within"  short:"e"  description:"timeframe in which to check expiration. Default: \"3m\".\n\t\t\t\tdays(d), weeks(w), months(m) and years(y) supported."`
	}
}

func NewExpiringCertificates(service expiringCertsService, logger logger) *ExpiringCerts {
	return &ExpiringCerts{
		api:    service,
		logger: logger,
	}
}

func (e *ExpiringCerts) Execute(args []string) error {
	if e.Options.ExpiresWithin == "" {
		e.Options.ExpiresWithin = "3m"
	}

	err := e.validateConfig()
	if err != nil {
		return err
	}

	e.logger.Println("Getting expiring certificates...")
	expiringCerts, err := e.api.ListExpiringCertificates(e.Options.ExpiresWithin)
	if err != nil {
		return fmt.Errorf("could not fetch expiring certificates: %s", err)
	}

	if len(expiringCerts) == 0 {
		e.logger.Printf(color.GreenString("[✓] No certificates are expiring in %s\n"), e.Options.ExpiresWithin)
		return nil
	}

	e.logger.Println(color.RedString("Found expiring certificates in the foundation:\n"))

	expiringCertsWithVariablePath, expiringCertsWithProductGUID := e.groupByLocation(expiringCerts)
	for location, certs := range expiringCertsWithVariablePath {
		e.logger.Printf(color.RedString("[X] %s", location))

		for _, cert := range certs {
			e.printExpiringCertInfo(cert)
		}
	}

	for location, productGUIDs := range expiringCertsWithProductGUID {
		e.logger.Printf(color.RedString("[X] %s", location))
		for guid, certs := range productGUIDs {
			e.logger.Printf(color.RedString("    %s:", guid))
			for _, cert := range certs {
				e.printExpiringCertInfo(cert)
			}
		}
	}

	return errors.New("found expiring certificates in the foundation")
}

func (e *ExpiringCerts) groupByLocation(certs []api.ExpiringCertificate) (map[string][]api.ExpiringCertificate, map[string]map[string][]api.ExpiringCertificate) {
	expiringCertsWithVariablePath := make(map[string][]api.ExpiringCertificate)
	expiringCertsWithProductGUID := make(map[string]map[string][]api.ExpiringCertificate)
	for _, cert := range certs {
		location := strings.Title(strings.Replace(cert.Location, "_", " ", -1))
		if cert.VariablePath != "" {
			expiringCertsWithVariablePath[location] = append(expiringCertsWithVariablePath[location], cert)
			continue
		}

		if expiringCertsWithProductGUID[location] == nil {
			expiringCertsWithProductGUID[location] = make(map[string][]api.ExpiringCertificate)
		}

		expiringCertsWithProductGUID[location][cert.ProductGUID] = append(expiringCertsWithProductGUID[location][cert.ProductGUID], cert)
	}

	return expiringCertsWithVariablePath, expiringCertsWithProductGUID
}

func (e *ExpiringCerts) printExpiringCertInfo(cert api.ExpiringCertificate) {
	expiringStr := "expiring"
	if time.Now().After(cert.ValidUntil) {
		expiringStr = "expired"
	}

	validUntil := cert.ValidUntil.Format(time.RFC822)

	if cert.VariablePath != "" {
		e.logger.Printf(color.RedString("    %s: %s on %s"), cert.VariablePath, expiringStr, validUntil)
		return
	}

	e.logger.Printf(color.RedString("        %s: %s on %s"), cert.PropertyReference, expiringStr, validUntil)
}

func (e ExpiringCerts) validateConfig() error {
	matched, err := regexp.MatchString("^[1-9]+\\d*[dwmy]$", e.Options.ExpiresWithin)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("only d,w,m, or y are supported. Default is \"3m\"")
	}
	return nil
}
