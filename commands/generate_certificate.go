package commands

import (
	"fmt"
	"github.com/pivotal-cf/om/api"
	"strings"

	"github.com/pivotal-cf/jhanda"
)

type GenerateCertificate struct {
	service generateCertificateService
	logger  logger
	Options struct {
		Domains string `long:"domains" short:"d" required:"true" description:"domains to generate certificates, delimited by comma, can include wildcard domains"`
	}
}

//go:generate counterfeiter -o ./fakes/generate_certificate_service.go --fake-name GenerateCertificateService . generateCertificateService
type generateCertificateService interface {
	GenerateCertificate(domains api.DomainsInput) (string, error)
}

func NewGenerateCertificate(service generateCertificateService, logger logger) GenerateCertificate {
	return GenerateCertificate{service: service, logger: logger}
}

func (g GenerateCertificate) Execute(args []string) error {
	if _, err := jhanda.Parse(&g.Options, args); err != nil {
		return fmt.Errorf("could not parse generate-certificate flags: %s", err)
	}

	domains := strings.Split(g.Options.Domains, ",")
	for i, domain := range domains {
		domains[i] = strings.TrimSpace(domain)
	}

	output, err := g.service.GenerateCertificate(api.DomainsInput{
		Domains: domains,
	})

	if err != nil {
		return err
	}

	g.logger.Printf(output)
	return nil
}

func (g GenerateCertificate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command generates a new RSA public/private certificate signed by Ops Manager’s root CA certificate",
		ShortDescription: "generates a new certificate signed by Ops Manager's root CA",
		Flags:            g.Options,
	}
}
