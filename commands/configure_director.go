package commands

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ConfigureDirector struct {
	service directorService
	logger  logger
	Options struct {
		AZConfiguration       string `short:"a" long:"az-configuration" description:"configures network availability zones"`
		NetworksConfiguration string `short:"n" long:"networks-configuration" description:"configures networks for the bosh director"`
		NetworkAssignment     string `short:"na" long:"network-assignment" description:"assigns networks and AZs"`
		DirectorConfiguration string `short:"d" long:"director-configuration" description:"properties for director configuration"`
		IAASConfiguration     string `short:"i" long:"iaas-configuration" description:"iaas specific JSON configuration for the bosh director"`
		SecurityConfiguration string `short:"s" long:"security-configuration" decription:"security configuration properties for directory"`
		SyslogConfiguration   string `short:"l" long:"syslog-configuration" decription:"syslog configuration properties for directory"`
	}
}

//go:generate counterfeiter -o ./fakes/director_service.go --fake-name DirectorService . directorService

type directorService interface {
	AZConfiguration(api.AZConfiguration) error
	NetworksConfiguration(json.RawMessage) error
	NetworkAndAZ(api.NetworkAndAZConfiguration) error
	Properties(api.DirectorProperties) error
}

func NewConfigureDirector(service directorService, logger logger) ConfigureDirector {
	return ConfigureDirector{service: service, logger: logger}
}

func (c ConfigureDirector) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

	c.logger.Printf("started configuring director options for bosh tile")

	err := c.service.Properties(api.DirectorProperties{
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

		err = c.service.AZConfiguration(api.AZConfiguration{
			AvailabilityZones: json.RawMessage(c.Options.AZConfiguration),
		})
		if err != nil {
			return fmt.Errorf("availability zones configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring availability zone options for bosh tile")
	}

	if c.Options.NetworksConfiguration != "" {
		c.logger.Printf("started configuring network options for bosh tile")

		err = c.service.NetworksConfiguration(json.RawMessage(c.Options.NetworksConfiguration))
		if err != nil {
			return fmt.Errorf("networks configuration could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network options for bosh tile")
	}

	if c.Options.NetworkAssignment != "" {
		c.logger.Printf("started configuring network assignment options for bosh tile")

		err = c.service.NetworkAndAZ(api.NetworkAndAZConfiguration{
			NetworkAZ: json.RawMessage(c.Options.NetworkAssignment),
		})
		if err != nil {
			return fmt.Errorf("network and AZs could not be applied: %s", err)
		}

		c.logger.Printf("finished configuring network assignment options for bosh tile")
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
