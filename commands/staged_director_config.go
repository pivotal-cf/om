package commands

import (
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"fmt"
	"gopkg.in/yaml.v2"
)

type StagedDirectorConfig struct {
	logger  logger
	service stagedDirectorConfigService
	Options struct{}
}

//go:generate counterfeiter -o ./fakes/staged_director_config_service.go --fake-name StagedDirectorConfigService . stagedDirectorConfigService
type stagedDirectorConfigService interface {
	GetStagedProductByName(product string) (api.StagedProductsFindOutput, error)
	GetStagedProductProperties(product string) (map[string]api.ResponseProperty, error)
	GetStagedDirectorAZ() ([]interface{}, error) // /api/v0/staged/director/network_and_az
	GetStagedDirectorProperties() (map[string]interface{}, error)// /api/v0/staged/director/properties
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

	azConfiguration, err := ec.service.GetStagedDirectorAZ()

	directorProperties, err := ec.service.GetStagedDirectorProperties()


	config := map[string]interface{}{}
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

	err = yaml.Unmarshal([]byte(str), &config)
	if err != nil {
		return err
	}

	config["az-configuration"] = azConfiguration
	config["director-configuration"] = directorProperties

	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	ec.logger.Println(string(configYaml))

	return nil
}
