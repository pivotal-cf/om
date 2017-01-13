package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

const (
	iaasConfigurationPath              = "/infrastructure/iaas_configuration/edit"
	directorConfigurationPath          = "/infrastructure/director_configuration/edit"
	securityConfigurationPath          = "/infrastructure/security_tokens/edit"
	availabilityZonesConfigurationPath = "/infrastructure/availability_zones/edit"
	networksConfigurationPath          = "/infrastructure/networks/edit"
	networkAssignmentPath              = "/infrastructure/director/az_and_network_assignment/edit"
)

type ConfigureBosh struct {
	service boshFormService
	logger  logger
	Options struct {
		IaaSConfiguration              string `short:"i"  long:"iaas-configuration"  description:"iaas specific JSON configuration for the bosh director"`
		DirectorConfiguration          string `short:"d"  long:"director-configuration"  description:"director-specific JSON configuration for the bosh director"`
		SecurityConfiguration          string `short:"s"  long:"security-configuration"  description:"security-specific JSON configuration for the bosh director"`
		AvailabilityZonesConfiguration string `short:"a"  long:"az-configuration"  description:"availability zones JSON configuration for the bosh director"`
		NetworksConfiguration          string `short:"n"  long:"networks-configuration"  description:"complete network configuration for the bosh director"`
		NetworkAssignment              string `short:"na"  long:"network-assignment"  description:"choose existing network and availability zone to deploy bosh director into"`
	}
}

//go:generate counterfeiter -o ./fakes/bosh_form_service.go --fake-name BoshFormService . boshFormService
type boshFormService interface {
	GetForm(path string) (api.Form, error)
	PostForm(api.PostFormInput) error
	AvailabilityZones() (map[string]string, error)
	Networks() (map[string]string, error)
}

func NewConfigureBosh(s boshFormService, l logger) ConfigureBosh {
	return ConfigureBosh{service: s, logger: l}
}

func (c ConfigureBosh) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return err
	}

	if c.Options.IaaSConfiguration != "" {
		c.logger.Printf("configuring iaas specific options for bosh tile")

		config, err := c.configureForm(c.Options.IaaSConfiguration)
		if err != nil {
			return err
		}

		err = c.postForm(iaasConfigurationPath, config)
		if err != nil {
			return err
		}
	}

	if c.Options.DirectorConfiguration != "" {
		c.logger.Printf("configuring director options for bosh tile")

		config, err := c.configureForm(c.Options.DirectorConfiguration)
		if err != nil {
			return err
		}

		err = c.postForm(directorConfigurationPath, config)
		if err != nil {
			return err
		}
	}

	if c.Options.AvailabilityZonesConfiguration != "" {
		c.logger.Printf("configuring availability zones for bosh tile")

		config, err := c.configureForm(c.Options.AvailabilityZonesConfiguration)
		if err != nil {
			return err
		}

		for _, az := range config.AvailabilityZonesConfiguration.AvailabilityZones {
			if az.Cluster == "" {
				config.AvailabilityZonesConfiguration.Names = append(config.AvailabilityZonesConfiguration.Names, az.Name)
			} else {
				config.AvailabilityZonesConfiguration.VSphereNames = append(config.AvailabilityZonesConfiguration.VSphereNames, az.Name)
				config.AvailabilityZonesConfiguration.Clusters = append(config.AvailabilityZonesConfiguration.Clusters, az.Cluster)
				config.AvailabilityZonesConfiguration.ResourcePools = append(config.AvailabilityZonesConfiguration.ResourcePools, az.ResourcePool)
			}
		}
		config.AvailabilityZonesConfiguration.AvailabilityZones = nil
		err = c.postForm(availabilityZonesConfigurationPath, config)
		if err != nil {
			return err
		}
	}

	if c.Options.NetworksConfiguration != "" {
		c.logger.Printf("configuring network options for bosh tile")
		if err != nil {
			panic(err)
		}

		err = c.configureNetworkForm(networksConfigurationPath, c.Options.NetworksConfiguration)
		if err != nil {
			return err
		}
	}

	if c.Options.NetworkAssignment != "" {
		c.logger.Printf("assigning az and networks for bosh tile")

		config, err := c.configureForm(c.Options.NetworkAssignment)
		if err != nil {
			return err
		}

		networks, err := c.service.Networks()
		if err != nil {
			return err
		}
		config.NetworkGUID = networks[config.UserProvidedNetworkName]

		availabilityZones, err := c.AZMap()
		if err != nil {
			return err
		}

		if azGUID, ok := availabilityZones[config.UserProvidedAZName]; ok {
			config.AZGUID = azGUID
		} else {
			config.AZGUID = "null-az"
		}

		err = c.postForm(networkAssignmentPath, config)
		if err != nil {
			return err
		}
	}

	if c.Options.SecurityConfiguration != "" {
		c.logger.Printf("configuring security options for bosh tile")

		config, err := c.configureForm(c.Options.SecurityConfiguration)
		if err != nil {
			return err
		}

		err = c.postForm(securityConfigurationPath, config)
		if err != nil {
			return err
		}
	}

	c.logger.Printf("finished configuring bosh tile")
	return nil
}

func (c ConfigureBosh) configureForm(configuration string) (BoshConfiguration, error) {
	var initialConfig BoshConfiguration

	err := json.NewDecoder(strings.NewReader(configuration)).Decode(&initialConfig)
	if err != nil {
		return BoshConfiguration{}, fmt.Errorf("could not decode json: %s", err)
	}

	return initialConfig, nil
}

func (c ConfigureBosh) postForm(path string, initialConfig BoshConfiguration) error {
	form, err := c.service.GetForm(path)
	if err != nil {
		return fmt.Errorf("could not fetch form: %s", err)
	}

	initialConfig.AuthenticityToken = form.AuthenticityToken
	initialConfig.Method = form.RailsMethod

	formValues, err := query.Values(initialConfig)
	if err != nil {
		return err // cannot be tested
	}

	err = c.service.PostForm(api.PostFormInput{Form: form, EncodedPayload: formValues.Encode()})
	if err != nil {
		return fmt.Errorf("tile failed to configure: %s", err)
	}

	return nil
}

func (c ConfigureBosh) AZMap() (map[string]string, error) {
	if c.Options.AvailabilityZonesConfiguration != "" || strings.Contains(c.Options.NetworkAssignment, `"singleton_availability_zone"`) {
		return c.service.AvailabilityZones()
	}
	return map[string]string{}, nil
}

func (c ConfigureBosh) configureNetworkForm(path, configuration string) error {
	form, err := c.service.GetForm(path)
	if err != nil {
		return fmt.Errorf("could not fetch form: %s", err)
	}

	var initialConfig NetworksConfiguration
	err = json.NewDecoder(strings.NewReader(configuration)).Decode(&initialConfig)
	if err != nil {
		return fmt.Errorf("could not decode json: %s", err)
	}

	azMap, err := c.AZMap()
	if err != nil {
		return fmt.Errorf("could not fetch availability zones: %s", err)
	}

	for n, network := range initialConfig.Networks {
		for s, subnet := range network.Subnets {
			if len(subnet.AvailabilityZones) > 0 {
				for _, azName := range subnet.AvailabilityZones {
					if azGuid, ok := azMap[azName]; ok {
						initialConfig.Networks[n].Subnets[s].AvailabilityZoneGUIDs = append(initialConfig.Networks[n].Subnets[s].AvailabilityZoneGUIDs, azGuid)
					}
				}
			} else {
				initialConfig.Networks[n].Subnets[s].AvailabilityZoneGUIDs = append(initialConfig.Networks[n].Subnets[s].AvailabilityZoneGUIDs, "null-az")
			}
		}
	}

	initialConfig.AuthenticityToken = form.AuthenticityToken
	initialConfig.Method = form.RailsMethod

	formValues, err := query.Values(initialConfig)
	if err != nil {
		return err // cannot be tested
	}

	err = c.service.PostForm(api.PostFormInput{Form: form, EncodedPayload: formValues.Encode()})
	if err != nil {
		return fmt.Errorf("tile failed to configure: %s", err)
	}

	return nil
}

func (c ConfigureBosh) Usage() Usage {
	return Usage{
		Description:      "configures the bosh director that is deployed by the Ops Manager",
		ShortDescription: "configures Ops Manager deployed bosh director",
		Flags:            c.Options,
	}
}
