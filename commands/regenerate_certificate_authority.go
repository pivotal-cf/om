package commands

type RegenerateCertificateAuthority struct {
	service certificateAuthoritiesRegenerator
	logger  logger
}

//go:generate counterfeiter -o ./fakes/certificate_authority_regenerator.go --fake-name CertificateAuthorityGenerator . certificateAuthorityRegenerator
type certificateAuthoritiesRegenerator interface {
	Regenerate() error
}

func NewRegenerateCertificateAuthority(service certificateAuthoritiesRegenerator, logger logger) RegenerateCertificateAuthority {
	return RegenerateCertificateAuthority{service: service, logger: logger}
}
