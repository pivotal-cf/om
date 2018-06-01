package commands

import "github.com/pivotal-cf/jhanda"

type StagedDirectorConfig struct {
	logger  logger
	service stagedConfigService
	Options struct {
		Product            string `long:"product-name" short:"p" required:"true" description:"name of product"`
		IncludeCredentials bool   `short:"c" long:"include-credentials" description:"include credentials. note: requires product to have been deployed"`
	}
}

func NewStagedDirectorConfig(service stagedConfigService, logger logger) StagedDirectorConfig {
	return StagedDirectorConfig{
		logger:  logger,
		service: service,
	}
}

func (ec StagedDirectorConfig) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a config from a staged director that can be passed in to om configure-director",
		ShortDescription: "**EXPERIMENTAL** generates a config from a staged director",
		Flags:            ec.Options,
	}
}

func (ec StagedDirectorConfig) Execute(args []string) error {
	str := `---
az-configuration:
- name: some-az
director-configuration:
 max_threads: 5
iaas-configuration:
 iaas_specific_key: some-value
network-assignment:
 network:
   name: some-network
networks-configuration:
 networks:
 - network: network-1
resource-configuration:
 compilation:
   instance_type:
     id: m4.xlarge
security-configuration:
 trusted_certificates: some-certificate
syslog-configuration:
 syslogconfig: awesome
`

	ec.logger.Println(string(str))

	return nil
}
