package commands

import (
	"github.com/pivotal-cf/om/api"
)

type InstallationLog struct {
	service installationLogService
	logger  logger
	Options struct {
		Id int `long:"id" required:"true" description:"id of the installation to retrieve logs for"`
	}
}

//counterfeiter:generate -o ./fakes/installation_log_service.go --fake-name InstallationLogService . installationLogService
type installationLogService interface {
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
}

func NewInstallationLog(service installationLogService, logger logger) *InstallationLog {
	return &InstallationLog{
		service: service,
		logger:  logger,
	}
}

func (i InstallationLog) Execute(args []string) error {
	output, err := i.service.GetInstallationLogs(i.Options.Id)
	if err != nil {
		return err
	}
	i.logger.Print(output.Logs)
	return nil
}
