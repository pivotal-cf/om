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
		IncludePlaceholders bool `long:"include-placeholders" short:"r" description:"Replace obscured credentials to interpolatable placeholders.\n\t\t\t\t    To include credentials hidden by OpsMan, use with \"--no-redact\""`
		NoRedact            bool `long:"no-redact" description:"Redact IaaS values from director configuration"`
	}
}

//go:generate counterfeiter -o ./fakes/staged_director_config_service.go --fake-name StagedDirectorConfigService . stagedDirectorConfigService
type stagedDirectorConfigService interface {
	GetStagedDirectorProperties(bool) (map[string]interface{}, error)
	GetStagedDirectorIaasConfigurations(bool) (map[string][]map[string]interface{}, error)
	GetStagedDirectorAvailabilityZones() (api.AvailabilityZonesOutput, error)
	GetStagedDirectorNetworks() (api.NetworksConfigurationOutput, error)

	GetStagedProductByName(productName string) (api.StagedProductsFindOutput, error)
	GetStagedProductNetworksAndAZs(productGUID string) (map[string]interface{}, error)

	ListStagedProductJobs(productGUID string) (map[string]string, error)
	GetStagedProductJobResourceConfig(productGUID, jobGUID string) (api.JobProperties, error)

	ListStagedVMExtensions() ([]api.VMExtension, error)
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

	properties, err := ec.service.GetStagedDirectorProperties(!ec.Options.NoRedact)
	if err != nil {
		return err
	}

	multiIaasConfigs, err := ec.service.GetStagedDirectorIaasConfigurations(!ec.Options.NoRedact)
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

	vmExtensions, err := ec.service.ListStagedVMExtensions()
	if err != nil {
		return err
	}

	config := map[string]interface{}{}
	if azs.AvailabilityZones != nil {
		config["az-configuration"] = azs.AvailabilityZones
	}

	if multiIaasConfigs != nil {
		config["iaas-configurations"] = multiIaasConfigs["iaas_configurations"]
		if _, ok := properties["iaas_configuration"]; ok {
			delete(properties, "iaas_configuration")
		}
	}

	config["properties-configuration"] = properties
	config["network-assignment"] = assignedNetworkAZ
	config["networks-configuration"] = networks
	config["vmextensions-configuration"] = vmExtensions

	resourceConfigs := map[string]api.JobProperties{}
	for name, jobGUID := range jobs {
		resourceConfig, err := ec.service.GetStagedProductJobResourceConfig(directorGUID, jobGUID)
		if err != nil {
			return err
		}
		resourceConfigs[name] = resourceConfig
	}
	config["resource-configuration"] = resourceConfigs

	if !ec.Options.NoRedact && !ec.Options.IncludePlaceholders {
		if _, ok := config["properties-configuration"].(map[string]interface{})["iaas_configuration"]; ok {
			delete(config["properties-configuration"].(map[string]interface{}), "iaas_configuration")
		}

		if _, ok := config["iaas-configurations"]; ok {
			delete(config, "iaas-configurations")
		}
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
		return ec.handleTypedMap(prefix, typedValue)
	case map[interface{}]interface{}:
		return ec.handleUntypedMap(prefix, typedValue)
	case map[string]map[string]interface{}:
		return ec.handleMapOfMaps(prefix, typedValue)
	case []map[string]interface{}:
		return ec.handleSliceOfMaps(prefix, typedValue)
	case []interface{}:
		return ec.handleSlice(prefix, typedValue)
	case string, int, bool, nil:
		if strings.Contains(prefix, "iaas_configuration") {
			if ec.Options.IncludePlaceholders {
				return "((" + prefix + "))", nil
			}
		}

		if strings.Contains(prefix, "iaas-configurations") {
			if ec.Options.IncludePlaceholders {
				return "((" + prefix + "))", nil
			}
		}

		for _, filter := range filters {
			if strings.Contains(keyName, filter) {
				if ec.Options.IncludePlaceholders {
					return "((" + prefix + "))", nil
				}
				if ec.Options.NoRedact {
					return value, nil
				}
				return nil, nil
			}
		}
	}

	return value, nil
}

func (ec StagedDirectorConfig) handleTypedMap(prefix string, value map[string]interface{}) (interface{}, error) {
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

func (ec StagedDirectorConfig) handleUntypedMap(prefix string, value map[interface{}]interface{}) (interface{}, error) {
	newValue := map[interface{}]interface{}{}

	for innerKey, innerVal := range value {
		switch typedValue := innerVal.(type) {
		case map[interface{}]interface{}:
			returnedVal, err := ec.handleUntypedMap(prefix+"_"+innerKey.(string), typedValue)

			if err != nil {
				return nil, err
			}
			if returnedVal != nil {
				newValue[innerKey] = returnedVal
			}
		case map[string]interface{}:
			returnedVal, err := ec.handleTypedMap(prefix+"_"+innerKey.(string), typedValue)

			if err != nil {
				return nil, err
			}
			if returnedVal != nil {
				newValue[innerKey] = returnedVal
			}
		case string, bool, int, nil:
			returnedVal, err := ec.filterSecrets(prefix+"_"+innerKey.(string), innerKey.(string), innerVal)

			if err != nil {
				return nil, err
			}
			if returnedVal != nil {
				newValue[innerKey] = returnedVal
			}
		case []interface{}:
			return ec.handleSlice(prefix, typedValue)
		}
	}
	return newValue, nil
}

func (ec StagedDirectorConfig) handleMapOfMaps(prefix string, value map[string]map[string]interface{}) (interface{}, error) {
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

func (ec StagedDirectorConfig) handleSliceOfMaps(prefix string, value []map[string]interface{}) (interface{}, error) {
	var newValue []map[string]interface{}
	for innerIndex, innerVal := range value {
		returnedVal, err := ec.filterSecrets(prefix+"_"+strconv.Itoa(innerIndex), "", innerVal)
		if err != nil {
			return nil, err
		}
		if returnedVal != nil {
			newValue = append(newValue, returnedVal.(map[string]interface{}))
		}
	}
	return newValue, nil
}
