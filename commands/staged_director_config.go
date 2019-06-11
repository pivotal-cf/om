package commands

import (
	"fmt"
	"log"
	"reflect"
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

	ListVMTypes() ([]api.VMType, error)
}

func NewStagedDirectorConfig(service stagedDirectorConfigService, logger logger) StagedDirectorConfig {
	return StagedDirectorConfig{
		logger:  logger,
		service: service,
	}
}

func (sdc StagedDirectorConfig) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a config from a staged director that can be passed in to om configure-director",
		ShortDescription: "**EXPERIMENTAL** generates a config from a staged director",
		Flags:            sdc.Options,
	}
}

func (sdc StagedDirectorConfig) Execute(args []string) error {
	if _, err := jhanda.Parse(&sdc.Options, args); err != nil {
		return fmt.Errorf("could not parse staged-config flags: %s", err)
	}

	stagedDirector, err := sdc.service.GetStagedProductByName("p-bosh")
	if err != nil {
		return err
	}

	directorGUID := stagedDirector.Product.GUID

	azs, err := sdc.service.GetStagedDirectorAvailabilityZones()
	if err != nil {
		return err
	}

	properties, err := sdc.service.GetStagedDirectorProperties(!sdc.Options.NoRedact)
	if err != nil {
		return err
	}

	multiIaasConfigs, err := sdc.service.GetStagedDirectorIaasConfigurations(!sdc.Options.NoRedact)
	if err != nil {
		return err
	}

	networks, err := sdc.service.GetStagedDirectorNetworks()
	if err != nil {
		return err
	}

	assignedNetworkAZ, err := sdc.service.GetStagedProductNetworksAndAZs(directorGUID)
	if err != nil {
		return err
	}

	jobs, err := sdc.service.ListStagedProductJobs(directorGUID)
	if err != nil {
		return err
	}

	vmExtensions, err := sdc.service.ListStagedVMExtensions()
	if err != nil {
		return err
	}

	vmTypes, err := sdc.service.ListVMTypes()
	if err != nil {
		return err
	}

	if len(vmTypes) > 0 && vmTypes[0].BuiltIn {
		vmTypes = []api.VMType{}
	}

	vmTypesConfig := VMTypesConfiguration{}

	if len(vmTypes) > 0 {
		outputVMTypes := make([]api.CreateVMType, len(vmTypes), len(vmTypes))

		for i := range vmTypes {
			outputVMTypes[i] = vmTypes[i].CreateVMType
		}

		vmTypesConfig.CustomTypesOnly = true
		vmTypesConfig.VMTypes = outputVMTypes
	}

	config := map[string]interface{}{}
	if azs.AvailabilityZones != nil {
		config["az-configuration"] = azs.AvailabilityZones
	}

	if multiIaasConfigs != nil {
		sdc.removePropertiesIAASConfig(config, multiIaasConfigs, properties)

		sdc.removeIAASConfigurationsGUID(config)
	}

	config["properties-configuration"] = properties
	config["network-assignment"] = assignedNetworkAZ
	config["networks-configuration"] = networks
	config["vmextensions-configuration"] = vmExtensions
	config["vmtypes-configuration"] = vmTypesConfig

	sdc.removePropertiesIAASConfigGUID(config)

	resourceConfigs, err := sdc.getResourceConfigs(jobs, directorGUID)
	if err != nil {
		return err
	}
	config["resource-configuration"] = resourceConfigs

	if !sdc.Options.NoRedact && !sdc.Options.IncludePlaceholders {
		sdc.removeAllIAASConfiguration(config)
	}

	for key, value := range config {
		returnedVal, err := sdc.filterSecrets(key, key, value)
		if err != nil {
			return err
		}
		if returnedVal != nil {
			config[key] = returnedVal
		}
	}

	configYaml, err := yaml.Marshal(config)
	if err != nil {
		log.Println("hi")
		return err
	}

	sdc.logger.Println(string(configYaml))
	return nil
}

func (sdc StagedDirectorConfig) removePropertiesIAASConfig(config map[string]interface{}, multiIaasConfigs map[string][]map[string]interface{}, properties map[string]interface{}) {
	config["iaas-configurations"] = multiIaasConfigs["iaas_configurations"]
	if _, ok := properties["iaas_configuration"]; ok {
		delete(properties, "iaas_configuration")
	}
}

func (sdc StagedDirectorConfig) removeIAASConfigurationsGUID(config map[string]interface{}) {
	for _, config := range config["iaas-configurations"].([]map[string]interface{}) {
		if _, ok := config["guid"]; ok {
			delete(config, "guid")
		}
	}
}

func (sdc StagedDirectorConfig) removePropertiesIAASConfigGUID(config map[string]interface{}) {
	if propertiesConfig, ok := config["properties-configuration"].(map[string]interface{}); ok {
		if iaasConfig, ok := propertiesConfig["iaas_configuration"]; ok {
			switch iaasConfig.(type) {
			case map[string]interface{}:
				delete(iaasConfig.(map[string]interface{}), "guid")
			case map[interface{}]interface{}:
				delete(iaasConfig.(map[interface{}]interface{}), "guid")
			}
		}
	}
}

func (sdc StagedDirectorConfig) getResourceConfigs(jobs map[string]string, directorGUID string) (map[string]api.JobProperties, error) {
	resourceConfigs := map[string]api.JobProperties{}

	for name, jobGUID := range jobs {
		resourceConfig, err := sdc.service.GetStagedProductJobResourceConfig(directorGUID, jobGUID)
		if err != nil {
			return nil, err
		}
		resourceConfigs[name] = resourceConfig
	}

	return resourceConfigs, nil
}

func (sdc StagedDirectorConfig) removeAllIAASConfiguration(config map[string]interface{}) {
	if _, ok := config["properties-configuration"].(map[string]interface{})["iaas_configuration"]; ok {
		delete(config["properties-configuration"].(map[string]interface{}), "iaas_configuration")
	}

	if _, ok := config["iaas-configurations"]; ok {
		delete(config, "iaas-configurations")
	}
}

func (sdc StagedDirectorConfig) filterSecrets(prefix string, keyName string, value interface{}) (interface{}, error) {
	filters := []string{"password", "user", "key"}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Map:
		elements := map[string]interface{}{}
		iter := v.MapRange()
		for iter.Next() {
			innerKey := fmt.Sprintf("%s", iter.Key())
			innerValue := iter.Value()
			returnedVal, err := sdc.filterSecrets(prefix+"_"+innerKey, innerKey, innerValue.Interface())

			if err != nil {
				return nil, err
			}
			if returnedVal != nil {
				elements[innerKey] = returnedVal
			}
		}
		return elements, nil
	case reflect.Slice:
		elements := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			returnedVal, err := sdc.filterSecrets(prefix+"_"+strconv.Itoa(i), "", v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			if returnedVal != nil {
				elements = append(elements, returnedVal)
			}
		}
		return elements, nil
	case reflect.String, reflect.Int, reflect.Bool:
		if strings.Contains(prefix, "iaas_configuration") {
			if sdc.Options.IncludePlaceholders {
				return "((" + prefix + "))", nil
			}
		}

		if strings.Contains(prefix, "iaas-configurations") {
			if sdc.Options.IncludePlaceholders {
				return "((" + prefix + "))", nil
			}
		}

		for _, filter := range filters {
			if strings.Contains(keyName, filter) {
				if sdc.Options.IncludePlaceholders {
					return "((" + prefix + "))", nil
				}
				if sdc.Options.NoRedact {
					return value, nil
				}
				return nil, nil
			}
		}
	}

	return value, nil
}
