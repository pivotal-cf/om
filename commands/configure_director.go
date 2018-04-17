package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureDirector struct {
	service configureDirectorService
	logger  logger
	Options struct {
		AZConfiguration       string `short:"a" long:"az-configuration" description:"configures network availability zones"`
		NetworksConfiguration string `short:"n" long:"networks-configuration" description:"configures networks for the bosh director"`
		NetworkAssignment     string `short:"na" long:"network-assignment" description:"assigns networks and AZs"`
		DirectorConfiguration string `short:"d" long:"director-configuration" description:"properties for director configuration"`
		IAASConfiguration     string `short:"i" long:"iaas-configuration" description:"iaas specific JSON configuration for the bosh director"`
		SecurityConfiguration string `short:"s" long:"security-configuration" decription:"security configuration properties for director"`
		SyslogConfiguration   string `short:"l" long:"syslog-configuration" decription:"syslog configuration properties for director"`
		ResourceConfiguration string `short:"r" long:"resource-configuration" decription:"resource configuration properties for director"`
	}
}

//go:generate counterfeiter -o ./fakes/configure_director_service.go --fake-name ConfigureDirectorService . configureDirectorService
type configureDirectorService interface {
	UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput) error
	UpdateStagedDirectorNetworks(json.RawMessage) error
	UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration) error
	UpdateStagedDirectorProperties(api.DirectorProperties) error
	ListStagedProductJobs(string) (map[string]string, error)
	GetStagedProductJobResourceConfig(string, string) (api.JobProperties, error)
	UpdateStagedProductJobResourceConfig(string, string, api.JobProperties) error
	GetStagedProductByName(name string) (api.StagedProductsFindOutput, error)
	GetStagedProductManifest(guid string) (manifest string, err error)
}

func NewConfigureDirector(service configureDirectorService, logger logger) ConfigureDirector {
	return ConfigureDirector{service: service, logger: logger}
}

func (c ConfigureDirector) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

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

	if c.Options.AZConfiguration != "" {
		c.logger.Printf("started configuring availability zone options for bosh tile")

		err = c.service.UpdateStagedDirectorAvailabilityZones(api.AvailabilityZoneInput{
			AvailabilityZones: json.RawMessage(c.Options.AZConfiguration),
		})
		if err != nil {
			return fmt.Errorf("availability zones configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring availability zone options for bosh tile")
	}

	if c.Options.NetworksConfiguration != "" {
		c.logger.Printf("started configuring network options for bosh tile")

		err = c.service.UpdateStagedDirectorNetworks(json.RawMessage(c.Options.NetworksConfiguration))
		if err != nil {
			return fmt.Errorf("networks configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network options for bosh tile")
	}

	if c.Options.NetworkAssignment != "" {
		c.logger.Printf("started configuring network assignment options for bosh tile")

		err = c.service.UpdateStagedDirectorNetworkAndAZ(api.NetworkAndAZConfiguration{
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
			return fmt.Errorf("could not decode resource-configuration json: %s", err)
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

	return nil
}

func (c ConfigureDirector) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command configures the director.",
		ShortDescription: "configures the director",
		Flags:            c.Options,
	}
}
