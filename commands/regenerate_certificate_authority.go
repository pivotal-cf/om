package commands

import "github.com/pivotal-cf/jhanda"

type RegenerateCertificateAuthority struct {
	service certificateAuthorityRegenerator
	logger  logger
}

//go:generate counterfeiter -o ./fakes/certificate_authority_regenerator.go --fake-name CertificateAuthorityRegenerator . certificateAuthorityRegenerator
type certificateAuthorityRegenerator interface {
	Regenerate() error
}

func NewRegenerateCertificateAuthority(service certificateAuthorityRegenerator, logger logger) RegenerateCertificateAuthority {
	return RegenerateCertificateAuthority{service: service, logger: logger}
}

func (r RegenerateCertificateAuthority) Execute(_ []string) error {
	err := r.service.Regenerate()
	if err != nil {
		return err
	}

	r.logger.Printf("Certificate authority regenerated.\n")

	return nil
}

func (r RegenerateCertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command regenerates certificate authority on the Ops Manager",
		ShortDescription: "regenerates a certificate authority on the Opsman",
	}
}
