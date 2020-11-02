package commands

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/presenters"
)

type CertificateAuthority struct {
	service   certificateAuthoritiesService
	presenter presenters.FormattedPresenter
	logger    logger
	Options   struct {
		ID      string `long:"id" description:"ID of certificate to display. Required if there is more than one certificate authority"`
		CertPEM bool   `long:"cert-pem" description:"Display the cert pem"`
		Format  string `long:"format" short:"f" default:"table" description:"Format to print as (options: table,json)"`
	}
}

func NewCertificateAuthority(certificateAuthoritiesService certificateAuthoritiesService, presenter presenters.FormattedPresenter, logger logger) *CertificateAuthority {
	return &CertificateAuthority{
		service:   certificateAuthoritiesService,
		presenter: presenter,
		logger:    logger,
	}
}

func (c CertificateAuthority) Execute(args []string) error {
	cas, err := c.service.ListCertificateAuthorities()
	if err != nil {
		return err
	}

	displayCA := api.CA{}
	if len(cas.CAs) == 1 && c.Options.ID == "" {
		displayCA = cas.CAs[0]
	} else {
		if len(cas.CAs) > 1 && c.Options.ID == "" {
			return errors.New("More than one certificate authority found. Please use --id flag to specify. IDs can be found using the certificate-authorities command")
		}
		for _, ca := range cas.CAs {
			if ca.GUID == c.Options.ID {
				displayCA = ca
				break
			}
		}
	}

	if !reflect.ValueOf(displayCA).IsZero() {
		if c.Options.CertPEM {
			c.logger.Println(displayCA.CertPEM)
		} else {
			c.presenter.SetFormat(c.Options.Format)
			c.presenter.PresentCertificateAuthority(displayCA)
		}
		return nil
	}

	return fmt.Errorf("could not find a certificate authority with ID: %q", c.Options.ID)
}
