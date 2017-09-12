package commands

import "strconv"

type GenerateCertificateAuthority struct {
	service     certificateAuthoritiesService
	tableWriter tableWriter
}

func NewGenerateCertificateAuthority(service certificateAuthoritiesService, tableWriter tableWriter) GenerateCertificateAuthority {
	return GenerateCertificateAuthority{service: service, tableWriter: tableWriter}
}

func (g GenerateCertificateAuthority) Execute(_ []string) error {
	ca, err := g.service.Generate()
	if err != nil {
		return err
	}

	g.tableWriter.SetAutoWrapText(false)
	g.tableWriter.SetHeader([]string{"id", "issuer", "active", "created on", "expired on", "certicate pem"})
	g.tableWriter.Append([]string{ca.GUID, ca.Issuer, strconv.FormatBool(ca.Active), ca.CreatedOn, ca.ExpiresOn, ca.CertPEM})
	g.tableWriter.Render()
	return nil
}

func (g GenerateCertificateAuthority) Usage() Usage {
	return Usage{
		Description:      "This authenticated command generates a certificate authority on the Ops Manager",
		ShortDescription: "generates a certificate authority on the Opsman",
	}
}
