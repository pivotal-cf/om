package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
	"github.com/pivotal-cf/om/api"
)

type ConfigureDirector struct {
	service directorService
	logger  logger
	Options struct {
		NetworkAssignment     string `short:"n" long:"network-assignment" description:"assigns networks and AZs"`
		DirectorConfiguration string `short:"d" long:"director-configuration" description:"properties for director configuration"`
		IAASConfiguration     string `short:"i" long:"iaas-configuration" description:"iaas specific JSON configuration for the bosh director"`
		SecurityConfiguration string `short:"s" long:"security-configuration" decription:"security configuration properties for directory"`
	}
}

//go:generate counterfeiter -o ./fakes/director_service.go --fake-name DirectorService . directorService
type directorService interface {
	NetworkAndAZ(api.NetworkAndAZConfiguration) error
	Properties(api.DirectorConfiguration) error
}

func NewConfigureDirector(service directorService, logger logger) ConfigureDirector {
	return ConfigureDirector{service: service, logger: logger}
}

func (c ConfigureDirector) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}

	if c.Options.NetworkAssignment != "" {
		var networkFields api.NetworkAndAZFields
		err = json.NewDecoder(strings.NewReader(c.Options.NetworkAssignment)).Decode(&networkFields)
		if err != nil {
			return err //not tested
		}

		c.logger.Printf("started configuring network assignment options for bosh tile")

		err = c.service.NetworkAndAZ(api.NetworkAndAZConfiguration{NetworkAZ: networkFields})
		if err != nil {
			return fmt.Errorf("network and AZs could not be applied: %s", err)
		}
	}

	c.logger.Printf("finished configuring network assignment options for bosh tile")

	c.logger.Printf("started configuring director options for bosh tile")

	err = c.service.Properties(api.DirectorConfiguration{
		DirectorConfiguration: json.RawMessage(c.Options.DirectorConfiguration),
		IAASConfiguration:     json.RawMessage(c.Options.IAASConfiguration),
		SecurityConfiguration: json.RawMessage(c.Options.SecurityConfiguration),
	})
	if err != nil {
		return fmt.Errorf("properties could not be applied: %s", err)
	}

	c.logger.Printf("finished configuring director options for bosh tile")

	return nil
}

func (c ConfigureDirector) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command configures the director.",
		ShortDescription: "configures the director",
		Flags:            c.Options,
	}
}
