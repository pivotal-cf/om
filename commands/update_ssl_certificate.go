package commands

import (
    "fmt"
    "os"

    "github.com/pivotal-cf/jhanda"
    "github.com/pivotal-cf/om/api"
)

type UpdateSSLCertificate struct {
    service updateSSLCertificateService
    logger  logger
    Options struct {
        CertPem    string   `long:"certificate-pem" required:"true" description:"certificate"`
        PrivateKey string   `long:"private-key-pem" required:"true" description:"private key"`
        ConfigFile string   `long:"config"                short:"c"                    description:"path to yml file for configuration (keys must match the following command line flags)"`
        VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV" experimental:"true"     description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
        VarsFile   []string `long:"vars-file"                                          description:"Load variables from a YAML file"`
        Vars       []string `long:"var"                                                description:"Load variable from the command line. Format: VAR=VAL"`
    }
}

//counterfeiter:generate -o ./fakes/update_ssl_certificate_service.go --fake-name UpdateSSLCertificateService . updateSSLCertificateService
type updateSSLCertificateService interface {
    UpdateSSLCertificate(api.SSLCertificateInput) error
}

func NewUpdateSSLCertificate(service updateSSLCertificateService, logger logger) UpdateSSLCertificate {
    return UpdateSSLCertificate{service: service, logger: logger}
}

func (c UpdateSSLCertificate) Execute(args []string) error {
    err := loadConfigFile(args, &c.Options, os.Environ)
    if err != nil {
        return fmt.Errorf("could not parse update-ssl-certificate flags: %s", err)
    }

    err = c.service.UpdateSSLCertificate(api.SSLCertificateInput{
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
        Description:      "This authenticated command updates the SSL Certificate on the Ops Manager with the given cert and key",
        ShortDescription: "updates the SSL Certificate on the Ops Manager",
        Flags:            c.Options,
    }
}
