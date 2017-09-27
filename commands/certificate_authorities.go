package commands

import (
	"strconv"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
)

type CertificateAuthorities struct {
	cas certificateAuthorityLister
	tw  tableWriter
}

//go:generate counterfeiter -o ./fakes/certificate_authority_lister.go --fake-name CertificateAuthorityLister . certificateAuthorityLister
type certificateAuthorityLister interface {
	List() (api.CertificateAuthoritiesOutput, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthorityLister, tableWriter tableWriter) CertificateAuthorities {
	return CertificateAuthorities{
		cas: certificateAuthoritiesService,
		tw:  tableWriter,
	}
}

func (c CertificateAuthorities) Execute(_ []string) error {
	casOutput, err := c.cas.List()
	if err != nil {
		return err
	}

	c.tw.SetAutoWrapText(false)
	c.tw.SetHeader([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"})

	for _, values := range casOutput.CAs {
		c.tw.Append([]string{values.GUID, values.Issuer, strconv.FormatBool(values.Active), values.CreatedOn, values.ExpiresOn, values.CertPEM})
	}

	c.tw.Render()

	return nil
}

func (c CertificateAuthorities) Usage() commands.Usage {
	return commands.Usage{
		Description:      "lists certificates managed by Ops Manager",
		ShortDescription: "lists certificates managed by Ops Manager",
	}
}
