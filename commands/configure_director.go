package commands

import (
	"encoding/json"
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"gopkg.in/yaml.v2"
	"sort"
	"strings"
)

type ConfigureDirector struct {
	environFunc func() []string
	service     configureDirectorService
	logger      logger
	Options     struct {
		ConfigFile string   `short:"c" long:"config" description:"path to yml file containing all config fields (see docs/configure-director/README.md for format)" required:"true"`
		VarsFile   []string `long:"vars-file"  description:"Load variables from a YAML file"`
		VarsEnv    []string `long:"vars-env"   description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		OpsFile    []string `long:"ops-file"  description:"YAML operations file"`
	}
}

type directorConfig struct {
	NetworkAssignment     interface{}            `yaml:"network-assignment"`
	AZConfiguration       interface{}            `yaml:"az-configuration"`
	NetworksConfiguration interface{}            `yaml:"networks-configuration"`
	DirectorConfigration  interface{}            `yaml:"director-configuration"`
	IaasConfiguration     interface{}            `yaml:"iaas-configuration"`
	SecurityConfiguration interface{}            `yaml:"security-configuration"`
	SyslogConfiguration   interface{}            `yaml:"syslog-configuration"`
	ResourceConfiguration map[string]interface{} `yaml:"resource-configuration"`
	VMExtensions          interface{}            `yaml:"vmextensions-configuration"`
	Field                 map[string]interface{} `yaml:",inline"`
}

//go:generate counterfeiter -o ./fakes/configure_director_service.go --fake-name ConfigureDirectorService . configureDirectorService
type configureDirectorService interface {
	CreateStagedVMExtension(api.CreateVMExtension) error
	DeleteVMExtension(name string) error
	GetStagedProductByName(name string) (api.StagedProductsFindOutput, error)
	GetStagedProductJobResourceConfig(string, string) (api.JobProperties, error)
	GetStagedProductManifest(guid string) (manifest string, err error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
	ListStagedProductJobs(string) (map[string]string, error)
	ListStagedVMExtensions() ([]api.VMExtension, error)
	UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput) error
	UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration) error
	UpdateStagedDirectorNetworks(api.NetworkInput) error
	UpdateStagedDirectorProperties(api.DirectorProperties) error
	UpdateStagedProductJobResourceConfig(string, string, api.JobProperties) error
}

func NewConfigureDirector(environFunc func() []string, service configureDirectorService, logger logger) ConfigureDirector {
	return ConfigureDirector{
		environFunc: environFunc,
		service:     service,
		logger:      logger,
	}
}

func (c ConfigureDirector) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command configures the director.",
		ShortDescription: "configures the director",
		Flags:            c.Options,
	}
}

func (c ConfigureDirector) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

	err := checkRunningInstallation(c.service.ListInstallations)
	if err != nil {
		return err
	}


	config, err := c.interpolateConfig()
	if err != nil {
		return err
	}

	err = c.validateConfig(config)
	if err != nil {
		return err
	}

	err = c.updateStagedDirectorProperties(config)
	if err != nil {
		return err
	}

	err = c.configureAvailabilityZones(config)
	if err != nil {
		return err
	}

	err = c.configureNetworksConfiguration(config)
	if err != nil {
		return err
	}

	err = c.configureNetworkAssignment(config)
	if err != nil {
		return err
	}

	err = c.configureResourceConfigurations(config)
	if err != nil {
		return err
	}

	err = c.configureVMExtensions(config)
	if err != nil {
		return err
	}

	return nil
}

func (c ConfigureDirector) interpolateConfig() (*directorConfig, error) {
	configContents, err := interpolate(interpolateOptions{
		templateFile: c.Options.ConfigFile,
		varsFiles:    c.Options.VarsFile,
		environFunc:  c.environFunc,
		varsEnvs:     c.Options.VarsEnv,
		opsFiles:     c.Options.OpsFile,
	}, "")
	if err != nil {
		return nil, err
	}

	var config directorConfig
	err = yaml.UnmarshalStrict(configContents, &config)
	if err != nil {
		return nil, fmt.Errorf("could not be parsed as valid configuration: %s: %s", c.Options.ConfigFile, err)
	}
	return &config, nil
}

func (c ConfigureDirector) validateConfig(config *directorConfig) error {
	if len(config.Field) > 0 {
		var unrecognizedKeys []string
		for key := range config.Field {
			unrecognizedKeys = append(unrecognizedKeys, key)
		}
		sort.Strings(unrecognizedKeys)

		return fmt.Errorf("the config file contains unrecognized keys: %s", strings.Join(unrecognizedKeys, ", "))
	}
	return nil
}

func (c ConfigureDirector) updateStagedDirectorProperties(config *directorConfig) error {
	if config.DirectorConfigration != nil || config.IaasConfiguration != nil || config.SecurityConfiguration != nil || config.SyslogConfiguration != nil {
		c.logger.Printf("started configuring director options for bosh tile")

		var err error
		var directorConfig, iaasConfig, securityConfig, syslogConfig string
		if config.DirectorConfigration != nil {
			directorConfig, err = getJSONProperties(config.DirectorConfigration)
			if err != nil {
				return err
			}
		}
		if config.IaasConfiguration != nil {
			iaasConfig, err = getJSONProperties(config.IaasConfiguration)
			if err != nil {
				return err
			}
		}
		if config.SecurityConfiguration != nil {
			securityConfig, err = getJSONProperties(config.SecurityConfiguration)
			if err != nil {
				return err
			}
		}
		if config.SyslogConfiguration != nil {
			syslogConfig, err = getJSONProperties(config.SyslogConfiguration)
			if err != nil {
				return err
			}
		}

		err = c.service.UpdateStagedDirectorProperties(api.DirectorProperties{
			DirectorConfiguration: json.RawMessage(directorConfig),
			IAASConfiguration:     json.RawMessage(iaasConfig),
			SecurityConfiguration: json.RawMessage(securityConfig),
			SyslogConfiguration:   json.RawMessage(syslogConfig),
		})

		if err != nil {
			return fmt.Errorf("properties could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring director options for bosh tile")
	}
	return nil
}

func (c ConfigureDirector) configureAvailabilityZones(config *directorConfig) error {
	if config.AZConfiguration != nil {
		c.logger.Printf("started configuring availability zone options for bosh tile")

		azs, err := getJSONProperties(config.AZConfiguration)
		if err != nil {
			return err
		}

		err = c.service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
			AvailabilityZones: json.RawMessage(azs),
		})
		if err != nil {
			return fmt.Errorf("availability zones configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring availability zone options for bosh tile")
	}
	return nil
}

func (c ConfigureDirector) configureNetworksConfiguration(config *directorConfig) error {
	if config.NetworksConfiguration != nil {
		c.logger.Printf("started configuring network options for bosh tile")

		networksConfiguration, err := getJSONProperties(config.NetworksConfiguration)
		if err != nil {
			return err
		}

		err = c.service.UpdateStagedDirectorNetworks(api.NetworkInput{
			Networks: json.RawMessage(networksConfiguration),
		})
		if err != nil {
			return fmt.Errorf("networks configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network options for bosh tile")
	}
	return nil
}

func (c ConfigureDirector) configureNetworkAssignment(config *directorConfig) error {
	if config.NetworkAssignment != nil {
		c.logger.Printf("started configuring network assignment options for bosh tile")

		networkAssignment, err := getJSONProperties(config.NetworkAssignment)
		if err != nil {
			return err
		}
		err = c.service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
			NetworkAZ: json.RawMessage(networkAssignment),
		})
		if err != nil {
			return fmt.Errorf("network and AZs could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network assignment options for bosh tile")
	}

	return nil
}

func (c ConfigureDirector) configureResourceConfigurations(config *directorConfig) error {
	if config.ResourceConfiguration != nil {
		c.logger.Printf("started configuring resource options for bosh tile")

		productGUID, err := c.getProductGUID()
		if err != nil {
			return err
		}

		jobNames := c.getUserProvidedJobNames(config)

		jobs, err := c.service.ListStagedProductJobs(productGUID)
		if err != nil {
			return fmt.Errorf("failed to fetch jobs: %s", err)
		}

		c.logger.Printf("applying resource configuration for the following jobs:")
		for _, jobName := range jobNames {
			err := c.configureResourceConfiguration(config, jobs, jobName, productGUID)
			if err != nil {
				return err
			}
		}

		c.logger.Printf("finished configuring resource options for bosh tile")
	}
	return nil
}

func (c ConfigureDirector) addNewExtensions(extensionsToDelete map[string]api.VMExtension, newExtensions interface{}) (map[string]api.VMExtension, error) {
	var newVMExtensions []api.VMExtension

	newExtensionBytes, err := getJSONProperties(newExtensions)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(newExtensionBytes), &newVMExtensions)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall vmextensions-configuration json: %s. Full Error: %s", newExtensions, err)
	}

	c.logger.Printf("applying vm-extensions configuration for the following:")
	for _, newExtension := range newVMExtensions {
		c.logger.Printf("\t%s", newExtension.Name)

		cloudProperties, err := json.Marshal(newExtension.CloudProperties)
		if err != nil {
			return nil, err
		}

		err = c.service.CreateStagedVMExtension(api.CreateVMExtension{
			Name:            newExtension.Name,
			CloudProperties: cloudProperties,
		})
		if err != nil {
			return nil, err
		}

		for name := range extensionsToDelete {
			if name == newExtension.Name {
				delete(extensionsToDelete, name)
			}
		}
	}
	return extensionsToDelete, nil
}

func (c ConfigureDirector) getExistingExtensions() (map[string]api.VMExtension, error) {
	existingVMExtensions, err := c.service.ListStagedVMExtensions()
	if err != nil {
		return nil, err
	}

	extensionsToDelete := make(map[string]api.VMExtension)
	for _, vmExtension := range existingVMExtensions {
		extensionsToDelete[vmExtension.Name] = vmExtension
	}

	return extensionsToDelete, nil
}

func (c ConfigureDirector) deleteExtensions(extensionsToDelete map[string]api.VMExtension) error {
	for _, extensionToDelete := range extensionsToDelete {
		c.logger.Printf("deleting vm extension %s", extensionToDelete.Name)
		err := c.service.DeleteVMExtension(extensionToDelete.Name)
		if err != nil {
			return err
		}
		c.logger.Printf("done deleting vm extension %s", extensionToDelete.Name)
	}
	return nil
}

func (c ConfigureDirector) configureVMExtensions(config *directorConfig) error {
	if config.VMExtensions != nil {
		c.logger.Printf("started configuring vm extensions")

		currentExtensions, err := c.getExistingExtensions()
		if err != nil {
			return err
		}

		extensionsToDelete, err := c.addNewExtensions(currentExtensions, config.VMExtensions)
		if err != nil {
			return err
		}

		err = c.deleteExtensions(extensionsToDelete)
		if err != nil {
			return err
		}

		c.logger.Printf("finished configuring vm extensions")
	}
	return nil
}

func (c ConfigureDirector) getProductGUID() (string, error) {
	findOutput, err := c.service.GetStagedProductByName("p-bosh")
	if err != nil {
		return "", fmt.Errorf("could not find staged product with name 'p-bosh': %s", err)
	}
	return findOutput.Product.GUID, nil
}

func (c ConfigureDirector) getUserProvidedJobNames(config *directorConfig) []string {
	var names []string

	for name, _ := range config.ResourceConfiguration {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (c ConfigureDirector) configureResourceConfiguration(config *directorConfig, jobs map[string]string, jobName string, productGUID string) error {
	c.logger.Printf("\t%s", jobName)
	jobGUID, ok := jobs[jobName]
	if !ok {
		return fmt.Errorf("product 'p-bosh' does not contain a job named '%s'", jobName)
	}

	jobProperties, err := c.service.GetStagedProductJobResourceConfig(productGUID, jobGUID)
	if err != nil {
		return fmt.Errorf("could not fetch existing job configuration for '%s': %s", jobName, err)
	}

	prop, err := getJSONProperties(config.ResourceConfiguration[jobName])
	if err != nil {
		return fmt.Errorf("could not unmarshall resource configuration: %v", err)
	}

	err = json.Unmarshal([]byte(prop), &jobProperties)
	if err != nil {
		return fmt.Errorf("could not decode resource-configuration json for job '%s': %s", jobName, err)
	}

	err = c.service.UpdateStagedProductJobResourceConfig(productGUID, jobGUID, jobProperties)
	if err != nil {
		return fmt.Errorf("failed to configure resources for '%s': %s", jobName, err)
	}

	return nil
}

func checkRunningInstallation(listInstallations func() ([]api.InstallationsServiceOutput, error)) error {
	installations, err := listInstallations()
	if err != nil {
		return fmt.Errorf("could not list installations: %s", err)
	}
	if len(installations) > 0 {
		if installations[0].Status == "running" {
			return fmt.Errorf("OpsManager does not allow configuration or staging changes while apply changes are running to prevent data loss for configuration and/or staging changes")
		}
	}
	return nil
}