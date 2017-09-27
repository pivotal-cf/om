package commands

import (
	"strconv"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/om/api"
)

type GenerateCertificateAuthority struct {
	service     certificateAuthorityGenerator
	tableWriter tableWriter
}

//go:generate counterfeiter -o ./fakes/certificate_authority_generator.go --fake-name CertificateAuthorityGenerator . certificateAuthorityGenerator
type certificateAuthorityGenerator interface {
	Generate() (api.CA, error)
}

func NewGenerateCertificateAuthority(service certificateAuthorityGenerator, tableWriter tableWriter) GenerateCertificateAuthority {
	return GenerateCertificateAuthority{service: service, tableWriter: tableWriter}
}

func (g GenerateCertificateAuthority) Execute(_ []string) error {
	ca, err := g.service.Generate()
	if err != nil {
		return err
	}

	g.tableWriter.SetAutoWrapText(false)
	g.tableWriter.SetHeader([]string{"id", "issuer", "active", "created on", "expires on", "certicate pem"})
	g.tableWriter.Append([]string{ca.GUID, ca.Issuer, strconv.FormatBool(ca.Active), ca.CreatedOn, ca.ExpiresOn, ca.CertPEM})
	g.tableWriter.Render()
	return nil
}

func (g GenerateCertificateAuthority) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command generates a certificate authority on the Ops Manager",
		ShortDescription: "generates a certificate authority on the Opsman",
	}
}
