package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthority struct {
	cas       certificateAuthoritiesService
	presenter presenters.Presenter
	logger    logger
	Options   struct {
		ID      string `long:"id"       required:"true" description:"ID of certificate to display"`
		CertPEM bool   `long:"cert-pem"                 description:"Display the cert pem"`
	}
}

func NewCertificateAuthority(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.Presenter, logger logger) CertificateAuthority {
	return CertificateAuthority{
		cas:       certificateAuthoritiesService,
		presenter: presenter,
		logger:    logger,
	}
}

func (c CertificateAuthority) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse certificate-authority flags: %s", err)
	}

	cas, err := c.cas.List()
	if err != nil {
		return err
	}

	for _, ca := range cas.CAs {
		if ca.GUID == c.Options.ID {
			if c.Options.CertPEM {
				c.logger.Println(ca.CertPEM)
			} else {
				c.presenter.PresentCertificateAuthority(ca)
			}
			return nil
		}
	}

	return fmt.Errorf("could not find a certificate authority with ID: %q", c.Options.ID)
}

func (CertificateAuthority) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "prints requested certificate authority",
		ShortDescription: "prints requested certificate authority",
	}
}
