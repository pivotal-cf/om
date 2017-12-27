package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type GenerateCertificate struct {
	service certificateGenerator
	logger  logger
	Options struct {
		Domains string `short:"d" long:"domains" description:"domains to generate certificates, delimited by comma, can include wildcard domains"`
	}
}

//go:generate counterfeiter -o ./fakes/certificate_generator.go --fake-name CertificateGenerator . certificateGenerator
type certificateGenerator interface {
	Generate(string) (string, error)
}

func NewGenerateCertificate(service certificateGenerator, logger logger) GenerateCertificate {
	return GenerateCertificate{service: service, logger: logger}
}

func (g GenerateCertificate) Execute(args []string) error {
	_, err := jhanda.Parse(&g.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse given flags: %s", err)
	}

	if g.Options.Domains == "" {
		return errors.New("error: domains is missing. Please see usage for more information.")
	}

	output, err := g.service.Generate(g.Options.Domains)
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
