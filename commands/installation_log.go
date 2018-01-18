package commands

import (
	"fmt"

	"github.com/pivotal-cf/jhanda"
)

type InstallationLog struct {
	service installationsService
	logger  logger
	Options struct {
		Id int `long:"id" required:"true" description:"id of the installation to retrieve logs for"`
	}
}

func NewInstallationLog(service installationsService, logger logger) InstallationLog {
	return InstallationLog{
		service: service,
		logger:  logger,
	}
}

func (i InstallationLog) Execute(args []string) error {
	if _, err := jhanda.Parse(&i.Options, args); err != nil {
		return fmt.Errorf("could not parse installation-log flags: %s", err)
	}

	output, err := i.service.Logs(i.Options.Id)
	if err != nil {
		return err
	}
	i.logger.Print(output.Logs)
	return nil
}

func (i InstallationLog) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command retrieves the logs for a given installation.",
		ShortDescription: "output installation logs",
		Flags:            i.Options,
	}
}
