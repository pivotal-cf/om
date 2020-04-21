package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type UpdateSSLCertificate struct {
	service     updateSSLCertificateService
	logger      logger
	environFunc func() []string
	Options     struct {
		CertPem    string   `long:"certificate-pem" required:"true" description:"certificate text"`
		PrivateKey string   `long:"private-key-pem" required:"true" description:"private key text"`
		ConfigFile string   `long:"config"                short:"c"                    description:"path to yml file for configuration (keys must match the following command line flags)"`
		VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV" experimental:"true"     description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
		VarsFile   []string `long:"vars-file"             short:"l"                    description:"load variables from a YAML file"`
		Vars       []string `long:"var"                   short:"v"                    description:"load variable from the command line. Format: VAR=VAL"`
	}
}

//counterfeiter:generate -o ./fakes/update_ssl_certificate_service.go --fake-name UpdateSSLCertificateService . updateSSLCertificateService
type updateSSLCertificateService interface {
	UpdateSSLCertificate(api.SSLCertificateSettings) error
}

func NewUpdateSSLCertificate(environFunc func() []string, service updateSSLCertificateService, logger logger) UpdateSSLCertificate {
	return UpdateSSLCertificate{environFunc: environFunc, service: service, logger: logger}
}

func (c UpdateSSLCertificate) Execute(args []string) error {
	err := loadConfigFile(args, &c.Options, c.environFunc)
	if err != nil {
		return fmt.Errorf("could not parse update-ssl-certificate flags: %s", err)
	}

	err = c.service.UpdateSSLCertificate(api.SSLCertificateSettings{
		CertPem:       c.Options.CertPem,
		PrivateKeyPem: c.Options.PrivateKey,
	})
	if err != nil {
		return err
	}

	c.logger.Printf("Successfully applied custom SSL Certificate.\n")
	c.logger.Printf("Please allow about 1 min for the new certificate to take effect.\n")

	return nil
}

func (c UpdateSSLCertificate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "***DEPRECATED*** This authenticated command updates the SSL Certificate on the Ops Manager with the given cert and key. Use configure-opsman instead.",
		ShortDescription: "**DEPRECATED** updates the SSL Certificate on the Ops Manager. Use configure-opsman instead.",
		Flags:            c.Options,
	}
}
