package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureOpsman struct {
	service     updateSSLCertificateService
	logger      logger
	environFunc func() []string
	Options     struct {
		ConfigFile string   `long:"config"                short:"c"                    description:"path to yml file for configuration"`
		VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV" experimental:"true"     description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
		VarsFile   []string `long:"vars-file"                                          description:"Load variables from a YAML file"`
		Vars       []string `long:"var"                                                description:"Load variable from the command line. Format: VAR=VAL"`
	}
}

type opsmanConfig struct {
	SSLCertificates interface{}            `yaml:"ssl-certificates"`
	Field           map[string]interface{} `yaml:",inline"`
}

//counterfeiter:generate -o ./fakes/configure_opsman_service.go --fake-name ConfigureOpsmanService . configureOpsmanService
type configureOpsmanService interface {
	UpdateSSLCertificate(api.SSLCertificateInput) error
}

func NewConfigureOpsman(environFunc func() []string, service updateSSLCertificateService, logger logger) ConfigureOpsman {
	return ConfigureOpsman{environFunc: environFunc, service: service, logger: logger}
}

func (c ConfigureOpsman) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-opsman flags: %s", err)
	}

	err := c.updateSSLCertificate()
	if err != nil {
		return err
	}
	//
	//c.logger.Printf("Successfully applied custom SSL Certificate.\n")
	//c.logger.Printf("Please allow about 1 min for the new certificate to take effect.\n")

	return nil
}

func (c ConfigureOpsman) updateSSLCertificate(config *opsmanConfig) error {
	//err := c.service.UpdateSSLCertificate(api.SSLCertificateInput{
	//	CertPem:       c.Options.CertPem,
	//	PrivateKeyPem: c.Options.PrivateKey,
	//})
}

func (c ConfigureOpsman) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command configures settings available on the \"Settings\" page in the Ops Manager UI",
		ShortDescription: "configures values present on the Ops Manager settings page",
		Flags:            c.Options,
	}
}
