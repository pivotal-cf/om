package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthority struct {
	cas       certificateAuthoritiesService
	presenter presenters.Presenter
	Options   struct {
		ID string `long:"id" description:"ID of certificate to display"`
	}
}

func NewCertificateAuthority(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.Presenter) CertificateAuthority {
	return CertificateAuthority{
		cas:       certificateAuthoritiesService,
		presenter: presenter,
	}
}

func (c CertificateAuthority) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse certificate-authority flags: %s", err)
	}

	cas, err := c.cas.List()
	if err != nil {
		return err
	}

	for _, ca := range cas.CAs {
		if ca.GUID == c.Options.ID {
			c.presenter.PresentCertificateAuthority(ca)
			return nil
		}
	}

	return fmt.Errorf("could not find a certificate authority with ID: %q", c.Options.ID)
}

func (CertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "prints requested certificate authority",
		ShortDescription: "prints requested certificate authority",
	}
}
