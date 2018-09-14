package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"gopkg.in/yaml.v2"
)

type ConfigureDirector struct {
	environFunc func() []string
	service     configureDirectorService
	logger      logger
	Options     struct {
		ConfigFile                string   `short:"c" long:"config" description:"path to yml file containing all config fields (see docs/configure-director/README.md for format)"`
		VarsFile                  []string `long:"vars-file"  description:"Load variables from a YAML file"`
		VarsEnv                   []string `long:"vars-env"   description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		OpsFile                   []string `long:"ops-file"  description:"YAML operations file"`
		AZConfiguration           string   `short:"a" long:"az-configuration" description:"configures network availability zones"`
		NetworksConfiguration     string   `short:"n" long:"networks-configuration" description:"configures networks for the bosh director"`
		NetworkAssignment         string   `short:"na" long:"network-assignment" description:"assigns networks and AZs"`
		DirectorConfiguration     string   `short:"d" long:"director-configuration" description:"properties for director configuration"`
		IAASConfiguration         string   `short:"i" long:"iaas-configuration" description:"iaas specific JSON configuration for the bosh director"`
		SecurityConfiguration     string   `short:"s" long:"security-configuration" decription:"security configuration properties for director"`
		SyslogConfiguration       string   `short:"l" long:"syslog-configuration" decription:"syslog configuration properties for director"`
		ResourceConfiguration     string   `short:"r" long:"resource-configuration" decription:"resource configuration properties for director"`
		VMExtensionsConfiguration string   `short:"v" long:"vmextensions-configuration" decription:"vm extensions configuration properties"`
	}
}

//go:generate counterfeiter -o ./fakes/configure_director_service.go --fake-name ConfigureDirectorService . configureDirectorService
type configureDirectorService interface {
	UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput) error
	UpdateStagedDirectorNetworks(api.NetworkInput) error
	UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration) error
	UpdateStagedDirectorProperties(api.DirectorProperties) error
	ListStagedProductJobs(string) (map[string]string, error)
	GetStagedProductJobResourceConfig(string, string) (api.JobProperties, error)
	UpdateStagedProductJobResourceConfig(string, string, api.JobProperties) error
	GetStagedProductByName(name string) (api.StagedProductsFindOutput, error)
	GetStagedProductManifest(guid string) (manifest string, err error)
	CreateStagedVMExtension(api.CreateVMExtension) error
	ListStagedVMExtensions() ([]api.VMExtension, error)
	DeleteVMExtension(name string) error
}

func NewConfigureDirector(environFunc func() []string, service configureDirectorService, logger logger) ConfigureDirector {
	return ConfigureDirector{
		environFunc: environFunc,
		service:     service,
		logger:      logger,
	}
}

func (c ConfigureDirector) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

	if c.Options.ConfigFile != "" {
		if c.Options.AZConfiguration != "" || c.Options.NetworksConfiguration != "" || c.Options.NetworkAssignment != "" || c.Options.DirectorConfiguration != "" || c.Options.IAASConfiguration != "" || c.Options.SecurityConfiguration != "" || c.Options.SyslogConfiguration != "" || c.Options.ResourceConfiguration != "" {
			return fmt.Errorf("config flag can not be passed with another configuration flags")
		}
		var config map[string]interface{}
		configContents, err := interpolate(interpolateOptions{
			templateFile: c.Options.ConfigFile,
			varsFiles:    c.Options.VarsFile,
			environFunc:  c.environFunc,
			varsEnvs:     c.Options.VarsEnv,
			opsFiles:     c.Options.OpsFile,
		})
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(configContents, &config)
		if err != nil {
			return fmt.Errorf("%s could not be parsed as valid configuration: %s", c.Options.ConfigFile, err)
		}

		if config["network-assignment"] != nil {
			c.Options.NetworkAssignment, err = getJSONProperties(config["network-assignment"])
			if err != nil {
				return err
			}
		}
		if config["az-configuration"] != nil {
			c.Options.AZConfiguration, err = getJSONProperties(config["az-configuration"])
			if err != nil {
				return err
			}
		}
		if config["networks-configuration"] != nil {
			c.Options.NetworksConfiguration, err = getJSONProperties(config["networks-configuration"])
			if err != nil {
				return err
			}
		}
		if config["director-configuration"] != nil {
			c.Options.DirectorConfiguration, err = getJSONProperties(config["director-configuration"])
			if err != nil {
				return err
			}
		}
		if config["iaas-configuration"] != nil {
			c.Options.IAASConfiguration, err = getJSONProperties(config["iaas-configuration"])
			if err != nil {
				return err
			}
		}
		if config["security-configuration"] != nil {
			c.Options.SecurityConfiguration, err = getJSONProperties(config["security-configuration"])
			if err != nil {
				return err
			}
		}
		if config["syslog-configuration"] != nil {
			c.Options.SyslogConfiguration, err = getJSONProperties(config["syslog-configuration"])
			if err != nil {
				return err
			}
		}
		if config["resource-configuration"] != nil {
			c.Options.ResourceConfiguration, err = getJSONProperties(config["resource-configuration"])
			if err != nil {
				return err
			}
		}
		if config["vmextensions-configuration"] != nil {
			c.Options.VMExtensionsConfiguration, err = getJSONProperties(config["vmextensions-configuration"])
			if err != nil {
				return err
			}
		}
	}

	if c.Options.DirectorConfiguration != "" || c.Options.IAASConfiguration != "" || c.Options.SecurityConfiguration != "" || c.Options.SyslogConfiguration != "" {
		c.logger.Printf("started configuring director options for bosh tile")

		err := c.service.UpdateStagedDirectorProperties(api.DirectorProperties{
			DirectorConfiguration: json.RawMessage(c.Options.DirectorConfiguration),
			IAASConfiguration:     json.RawMessage(c.Options.IAASConfiguration),
			SecurityConfiguration: json.RawMessage(c.Options.SecurityConfiguration),
			SyslogConfiguration:   json.RawMessage(c.Options.SyslogConfiguration),
		})

		if err != nil {
			return fmt.Errorf("properties could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring director options for bosh tile")
	}

	if c.Options.AZConfiguration != "" {
		c.logger.Printf("started configuring availability zone options for bosh tile")

		err := c.service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
			AvailabilityZones: json.RawMessage(c.Options.AZConfiguration),
		})
		if err != nil {
			return fmt.Errorf("availability zones configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring availability zone options for bosh tile")
	}

	if c.Options.NetworksConfiguration != "" {
		c.logger.Printf("started configuring network options for bosh tile")

		err := c.service.UpdateStagedDirectorNetworks(api.NetworkInput{
			Networks: json.RawMessage(c.Options.NetworksConfiguration),
		})
		if err != nil {
			return fmt.Errorf("networks configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network options for bosh tile")
	}

	if c.Options.NetworkAssignment != "" {
		c.logger.Printf("started configuring network assignment options for bosh tile")

		err := c.service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
			NetworkAZ: json.RawMessage(c.Options.NetworkAssignment),
		})
		if err != nil {
			return fmt.Errorf("network and AZs could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network assignment options for bosh tile")
	}

	if c.Options.ResourceConfiguration != "" {
		c.logger.Printf("started configuring resource options for bosh tile")

		findOutput, err := c.service.GetStagedProductByName("p-bosh")
		if err != nil {
			return fmt.Errorf("could not find staged product with name 'p-bosh': %s", err)
		}
		productGUID := findOutput.Product.GUID

		var userProvidedConfig map[string]json.RawMessage
		err = json.Unmarshal([]byte(c.Options.ResourceConfiguration), &userProvidedConfig)
		if err != nil {
			return fmt.Errorf("could not decode resource-configuration json: %s", c.Options.ResourceConfiguration)
		}

		jobs, err := c.service.ListStagedProductJobs(productGUID)
		if err != nil {
			return fmt.Errorf("failed to fetch jobs: %s", err)
		}

		var names []string
		for name, _ := range userProvidedConfig {
			names = append(names, name)
		}

		sort.Strings(names)

		c.logger.Printf("applying resource configuration for the following jobs:")
		for _, name := range names {
			c.logger.Printf("\t%s", name)
			jobGUID, ok := jobs[name]
			if !ok {
				return fmt.Errorf("product 'p-bosh' does not contain a job named '%s'", name)
			}

			jobProperties, err := c.service.GetStagedProductJobResourceConfig(productGUID, jobGUID)
			if err != nil {
				return fmt.Errorf("could not fetch existing job configuration for '%s': %s", name, err)
			}

			err = json.Unmarshal(userProvidedConfig[name], &jobProperties)
			if err != nil {
				return fmt.Errorf("could not decode resource-configuration json for job '%s': %s", name, err)
			}

			err = c.service.UpdateStagedProductJobResourceConfig(productGUID, jobGUID, jobProperties)
			if err != nil {
				return fmt.Errorf("failed to configure resources for '%s': %s", name, err)
			}
		}

		c.logger.Printf("finished configuring resource options for bosh tile")
	}

	if c.Options.VMExtensionsConfiguration != "" {
		c.logger.Printf("started configuring vm extensions")

		currentExtensions, err := c.getExistingExtensions()
		if err != nil {
			return err
		}

		extensionsToDelete, err := c.addNewExtensions(currentExtensions)
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

func (c ConfigureDirector) addNewExtensions(extensionsToDelete map[string]api.VMExtension) (map[string]api.VMExtension, error) {
	var newVMExtensions []api.VMExtension
	err := json.Unmarshal([]byte(c.Options.VMExtensionsConfiguration), &newVMExtensions)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall vmextensions-configuration json: %s. Full Error: %s", c.Options.VMExtensionsConfiguration, err)
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

func (c ConfigureDirector) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command configures the director.",
		ShortDescription: "configures the director",
		Flags:            c.Options,
	}
}
