package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"fmt"
)


type StagedDirectorConfig struct {
	logger  logger
	service stagedDirectorConfigService
	Options struct {}
}

//go:generate counterfeiter -o ./fakes/staged_director_config_service.go --fake-name StagedDirectorConfigService . stagedDirectorConfigService
type stagedDirectorConfigService interface {
	GetStagedProductByName(product string) (api.StagedProductsFindOutput, error)
	GetStagedProductProperties(product string) (map[string]api.ResponseProperty, error)
}


func NewStagedDirectorConfig(service stagedDirectorConfigService, logger logger) StagedDirectorConfig {
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
	if _, err := jhanda.Parse(&ec.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-config flags: %s", err)
	}

	findOutput, err := ec.service.GetStagedProductByName("p-bosh")
	if err != nil {
		return err
	}

	productGUID := findOutput.Product.GUID

	_, err = ec.service.GetStagedProductProperties(productGUID)
	if err != nil {
		return err
	}

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
