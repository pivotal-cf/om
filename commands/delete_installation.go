package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeleteInstallation struct {
	service      deleteInstallationService
	logger       logger
	logWriter    logWriter
	waitDuration time.Duration
}

//go:generate counterfeiter -o ./fakes/delete_installation_service.go --fake-name DeleteInstallationService . deleteInstallationService
type deleteInstallationService interface {
	DeleteInstallationAssetCollection() (api.InstallationsServiceOutput, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	GetInstallation(id int) (api.InstallationsServiceOutput, error)
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
}

func NewDeleteInstallation(service deleteInstallationService, logWriter logWriter, logger logger, waitDuration time.Duration) DeleteInstallation {
	return DeleteInstallation{
		service:      service,
		logger:       logger,
		logWriter:    logWriter,
		waitDuration: waitDuration,
	}
}

func (ac DeleteInstallation) Execute(args []string) error {
	installation, err := ac.service.RunningInstallation()

	if installation == (api.InstallationsServiceOutput{}) {
		ac.logger.Printf("attempting to delete the installation on the targeted Ops Manager")

		installation, err = ac.service.DeleteInstallationAssetCollection()
		if err != nil {
			return fmt.Errorf("failed to delete installation: %s", err)
		}

		if installation == (api.InstallationsServiceOutput{}) {
			ac.logger.Printf("no installation to delete")
			return nil
		}
	} else {
		ac.logger.Printf("found already running deletion...attempting to re-attach")
	}

	for {
		current, err := ac.service.GetInstallation(installation.ID)
		if err != nil {
			return fmt.Errorf("installation failed to get status: %s", err)
		}

		install, err := ac.service.GetInstallationLogs(installation.ID)
		if err != nil {
			return fmt.Errorf("installation failed to get logs: %s", err)
		}

		err = ac.logWriter.Flush(install.Logs)
		if err != nil {
			return fmt.Errorf("installation failed to flush logs: %s", err)
		}

		if current.Status == api.StatusSucceeded {
			return nil
		} else if current.Status == api.StatusFailed {
			return errors.New("deleting the installation was unsuccessful")
		}

		time.Sleep(ac.waitDuration)
	}
}

func (ac DeleteInstallation) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes all the products installed on the targeted Ops Manager.",
		ShortDescription: "deletes all the products on the Ops Manager targeted",
	}
}
