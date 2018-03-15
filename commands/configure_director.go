package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureDirector struct {
	directorService       directorService
	jobsService           jobsService
	stagedProductsService stagedProductsService
	logger                logger
	Options               struct {
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

//go:generate counterfeiter -o ./fakes/director_service.go --fake-name DirectorService . directorService

type directorService interface {
	AZConfiguration(api.AZConfiguration) error
	NetworksConfiguration(json.RawMessage) error
	NetworkAndAZ(api.NetworkAndAZConfiguration) error
	Properties(api.DirectorProperties) error
}

//go:generate counterfeiter -o ./fakes/jobs_service.go --fake-name JobsService . jobsService

type jobsService interface {
	Jobs(string) (map[string]string, error)
	GetExistingJobConfig(string, string) (api.JobProperties, error)
	ConfigureJob(string, string, api.JobProperties) error
}

//go:generate counterfeiter -o ./fakes/staged_products_service.go --fake-name StagedProductsService . stagedProductsService

type stagedProductsService interface {
	Find(name string) (api.StagedProductsFindOutput, error)
	Manifest(guid string) (manifest string, err error)
}

func NewConfigureDirector(directorService directorService, jobsService jobsService, stagedProductsService stagedProductsService, logger logger) ConfigureDirector {
	return ConfigureDirector{directorService: directorService, jobsService: jobsService, stagedProductsService: stagedProductsService, logger: logger}
}

func (c ConfigureDirector) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

	c.logger.Printf("started configuring director options for bosh tile")

	err := c.directorService.Properties(api.DirectorProperties{
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

		err = c.directorService.AZConfiguration(api.AZConfiguration{
			AvailabilityZones: json.RawMessage(c.Options.AZConfiguration),
		})
		if err != nil {
			return fmt.Errorf("availability zones configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring availability zone options for bosh tile")
	}

	if c.Options.NetworksConfiguration != "" {
		c.logger.Printf("started configuring network options for bosh tile")

		err = c.directorService.NetworksConfiguration(json.RawMessage(c.Options.NetworksConfiguration))
		if err != nil {
			return fmt.Errorf("networks configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network options for bosh tile")
	}

	if c.Options.NetworkAssignment != "" {
		c.logger.Printf("started configuring network assignment options for bosh tile")

		err = c.directorService.NetworkAndAZ(api.NetworkAndAZConfiguration{
			NetworkAZ: json.RawMessage(c.Options.NetworkAssignment),
		})
		if err != nil {
			return fmt.Errorf("network and AZs could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network assignment options for bosh tile")
	}

	if c.Options.ResourceConfiguration != "" {
		c.logger.Printf("started configuring resource options for bosh tile")

		findOutput, err := c.stagedProductsService.Find("p-bosh")
		if err != nil {
			return fmt.Errorf("could not find staged product with name 'p-bosh': %s", err)
		}
		productGUID := findOutput.Product.GUID

		var userProvidedConfig map[string]json.RawMessage
		err = json.Unmarshal([]byte(c.Options.ResourceConfiguration), &userProvidedConfig)
		if err != nil {
			return fmt.Errorf("could not decode resource-configuration json: %s", err)
		}

		jobs, err := c.jobsService.Jobs(productGUID)
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

			jobProperties, err := c.jobsService.GetExistingJobConfig(productGUID, jobGUID)
			if err != nil {
				return fmt.Errorf("could not fetch existing job configuration for '%s': %s", name, err)
			}

			err = json.Unmarshal(userProvidedConfig[name], &jobProperties)
			if err != nil {
				return fmt.Errorf("could not decode resource-configuration json for job '%s': %s", name, err)
			}

			err = c.jobsService.ConfigureJob(productGUID, jobGUID, jobProperties)
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
