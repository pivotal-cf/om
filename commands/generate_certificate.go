package commands

import (
	"github.com/pivotal-cf/om/api"
	"strings"
)

type GenerateCertificate struct {
	service generateCertificateService
	logger  logger
	Options struct {
		Domains []string `long:"domains" short:"d" required:"true" description:"domains to generate certificates, can include wildcard domains. Can be delimited by comma, specified in multiple flags, or both"`
	}
}

//counterfeiter:generate -o ./fakes/generate_certificate_service.go --fake-name GenerateCertificateService . generateCertificateService
type generateCertificateService interface {
	GenerateCertificate(domains api.DomainsInput) (string, error)
}

func NewGenerateCertificate(service generateCertificateService, logger logger) *GenerateCertificate {
	return &GenerateCertificate{service: service, logger: logger}
}

func (g GenerateCertificate) Execute(args []string) error {
	domains := make([]string, 0)

	for _, tmpDomain := range g.Options.Domains {
		tmpDomains := strings.Split(tmpDomain, ",")
		for _, domain := range tmpDomains {
			domains = append(domains, strings.TrimSpace(domain))
		}
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
