package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda/commands"
	"github.com/pivotal-cf/jhanda/flags"
)

type ConfigureDirector struct {
	service directorService
	Options struct {
		NetworkAssignment string `short:"n" long:"network-assignment" description:"assigns networks and AZs"`
	}
}

//go:generate counterfeiter -o ./fakes/director_service.go --fake-name DirectorService . directorService
type directorService interface {
	NetworkAndAZ(jsonBody string) error
}

func NewConfigureDirector(service directorService) ConfigureDirector {
	return ConfigureDirector{service: service}
}

func (c ConfigureDirector) Execute(args []string) error {
	_, err := flags.Parse(&c.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse configure-director flags: %s", err)
	}
	err = c.service.NetworkAndAZ(c.Options.NetworkAssignment)
	if err != nil {
		return fmt.Errorf("network and AZs couldn't be applied: %s", err)
	}
	return nil
}

func (c ConfigureDirector) Usage() commands.Usage {
	return commands.Usage{
		Description:      "This authenticated command configures the director.",
		ShortDescription: "configures the director",
		Flags:            c.Options,
	}
}
