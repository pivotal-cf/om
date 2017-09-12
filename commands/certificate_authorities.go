package commands

import (
	"strconv"
)

type CertificateAuthorities struct {
	cas certificateAuthoritiesService
	tw tableWriter
}

type CA struct {
	GUID string
	Issuer string
	CreatedOn string
	ExpiresOn string
	Active bool
	CertPEM string
}

type certificateAuthoritiesService interface {
	CertificateAuthorities() ([]CA, error)
}

func NewCertificateAuthorities(certificateAuthoritiesService certificateAuthoritiesService, tableWriter tableWriter) CertificateAuthorities {
	return CertificateAuthorities{
		cas: certificateAuthoritiesService,
		tw: tableWriter,
	}
}

func (c CertificateAuthorities) Execute(_ []string) error {
	cas, err := c.cas.CertificateAuthorities()
	if err != nil {
		return err
	}

	c.tw.SetHeader([]string{"id", "issuer", "active", "created on", "expired on", "certicate pem"})

	for _, values := range cas {
		c.tw.Append([]string{values.GUID, values.Issuer, strconv.FormatBool(values.Active), values.CreatedOn, values.ExpiresOn, values.CertPEM})
	}
	c.tw.Render()
	
	return nil
}