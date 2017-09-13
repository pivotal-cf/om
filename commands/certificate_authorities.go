package commands

import (
	"strconv"

	"github.com/pivotal-cf/om/api"
)

type CertificateAuthorities struct {
	cas certificateAuthoritiesService
	tw  tableWriter
}

//go:generate counterfeiter -o ./fakes/certificate_authorities_service.go --fake-name CertificateAuthoritiesService . certificateAuthoritiesService
type certificateAuthoritiesService interface {
	List() (api.CertificateAuthoritiesServiceOutput, error)
	Generate() (api.CA, error)
	Create(api.CertificateAuthorityBody) (api.CA, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthoritiesService, tableWriter tableWriter) CertificateAuthorities {
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
	c.tw.SetHeader([]string{"id", "issuer", "active", "created on", "expired on", "certicate pem"})

	for _, values := range casOutput.CAs {
		c.tw.Append([]string{values.GUID, values.Issuer, strconv.FormatBool(values.Active), values.CreatedOn, values.ExpiresOn, values.CertPEM})
	}

	c.tw.Render()

	return nil
}

func (c CertificateAuthorities) Usage() Usage {
	return Usage{
		Description:      "lists certificates managed by Ops Manager",
		ShortDescription: "lists certificates managed by Ops Manager",
	}
}
