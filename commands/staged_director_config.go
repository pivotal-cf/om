package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"gopkg.in/yaml.v2"
)

type StagedDirectorConfig struct {
	logger  logger
	service stagedDirectorConfigService
	Options struct {
		IncludeCredentials bool `short:"c" long:"include-credentials" description:"include credentials. note: requires product to have been deployed"`
		IncludePlaceholder bool `short:"r" long:"include-placeholder" description:"replace obscured credentials to interpolatable placeholder"`
	}
}

//go:generate counterfeiter -o ./fakes/staged_director_config_service.go --fake-name StagedDirectorConfigService . stagedDirectorConfigService
type stagedDirectorConfigService interface {
	GetStagedDirectorProperties() (map[string]map[string]interface{}, error)
	GetStagedDirectorAvailabilityZones() (api.AvailabilityZonesOutput, error)
	GetStagedDirectorNetworks() (api.NetworksConfigurationOutput, error)

	GetStagedProductByName(productName string) (api.StagedProductsFindOutput, error)
	GetStagedProductNetworksAndAZs(productGUID string) (map[string]interface{}, error)

	ListStagedProductJobs(productGUID string) (map[string]string, error)
	GetStagedProductJobResourceConfig(productGUID, jobGUID string) (api.JobProperties, error)
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

	stagedDirector, err := ec.service.GetStagedProductByName("p-bosh")
	if err != nil {
		return err
	}

	directorGUID := stagedDirector.Product.GUID

	azs, err := ec.service.GetStagedDirectorAvailabilityZones()
	if err != nil {
		return err
	}

	properties, err := ec.service.GetStagedDirectorProperties()
	if err != nil {
		return err
	}

	networks, err := ec.service.GetStagedDirectorNetworks()
	if err != nil {
		return err
	}

	assignedNetworkAZ, err := ec.service.GetStagedProductNetworksAndAZs(directorGUID)
	if err != nil {
		return err
	}

	jobs, err := ec.service.ListStagedProductJobs(directorGUID)
	if err != nil {
		return err
	}

	config := map[string]interface{}{}
	config["az-configuration"] = azs.AvailabilityZones
	config["director-configuration"] = properties["director_configuration"]
	config["iaas-configuration"] = properties["iaas_configuration"]
	config["syslog-configuration"] = properties["syslog_configuration"]
	config["security-configuration"] = properties["security_configuration"]
	config["network-assignment"] = assignedNetworkAZ
	config["networks-configuration"] = networks

	resourceConfigs := map[string]api.JobProperties{}
	for name, jobGUID := range jobs {
		resourceConfig, err := ec.service.GetStagedProductJobResourceConfig(directorGUID, jobGUID)
		if err != nil {
			return err
		}
		resourceConfigs[name] = resourceConfig
	}
	config["resource-configuration"] = resourceConfigs

	if !ec.Options.IncludeCredentials && !ec.Options.IncludePlaceholder {
		delete(config, "iaas-configuration")
	}

	for key, value := range config {
		returnedVal, err := ec.filterSecrets(key, key, value)
		if err != nil {
			return err
		}
		if returnedVal != nil {
			config[key] = returnedVal
		}
	}

	configYaml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	ec.logger.Println(string(configYaml))
	return nil
}

func (ec StagedDirectorConfig) filterSecrets(prefix string, keyName string, value interface{}) (interface{}, error) {
	filters := []string{"password", "user", "key"}
	switch typedValue := value.(type) {
	case map[string]interface{}:
		return ec.handleMap(prefix, typedValue)

	case []interface{}:
		return ec.handleSlice (prefix, typedValue)

	case string, nil:
		if strings.Contains(prefix, "iaas-configuration") {
			if ec.Options.IncludePlaceholder {
				return "((" + prefix + "))", nil
			}
		}

		for _, filter := range filters {
			if strings.Contains(keyName, filter) {
				if ec.Options.IncludePlaceholder {
					return "((" + prefix + "))", nil
				}
				if ec.Options.IncludeCredentials {
					return value, nil
				}
				return nil, nil
			}
		}
	}
	return value, nil
}

func (ec StagedDirectorConfig) handleMap(prefix string, value map[string]interface{}) (interface{}, error) {
	newValue := map[string]interface{}{}
	for innerKey, innerVal := range value {
		returnedVal, err := ec.filterSecrets(prefix+"_"+innerKey, innerKey, innerVal)

		if err != nil {
			return nil, err
		}
		if returnedVal != nil {
			newValue[innerKey] = returnedVal
		}
	}
	return newValue, nil
}

func (ec StagedDirectorConfig) handleSlice(prefix string, value []interface{}) (interface{}, error) {
	var newValue []interface{}
	for innerIndex, innerVal := range value {
		returnedVal, err := ec.filterSecrets(prefix+"_"+strconv.Itoa(innerIndex), "", innerVal)
		if err != nil {
			return nil, err
		}
		if returnedVal != nil {
			newValue = append(newValue, returnedVal)
		}
	}
	return newValue, nil
}
