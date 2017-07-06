package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/om/flags"
)

type InstallationLog struct {
	service installationsService
	logger  logger
	Options struct {
		Id int `short:"id" description:"id of the installation to retrieve logs for"`
	}
}

func NewInstallationLog(service installationsService, logger logger) InstallationLog {
	return InstallationLog{
		service: service,
		logger:  logger,
	}
}

func (i InstallationLog) Execute(args []string) error {
	_, err := flags.Parse(&i.Options, args)

	if err != nil {
		return fmt.Errorf("could not parse installation-log flags: %s", err)
	}

	if i.Options.Id == 0 {
		return errors.New("error: id is missing. Please see usage for more information.")
	}

	output, err := i.service.Logs(i.Options.Id)
	if err != nil {
		return err
	}
	i.logger.Print(output.Logs)
	return nil
}

func (i InstallationLog) Usage() Usage {
	return Usage{
		Description:      "This authenticated command retrieves the logs for a given installation.",
		ShortDescription: "output installation logs",
		Flags:            i.Options,
	}
}
