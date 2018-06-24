package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
	"errors"
)

type DeleteInstallation struct {
	service deleteInstallationService
	logger  logger
}

//go:generate counterfeiter -o ./fakes/delete_installation_service.go --fake-name DeleteInstallationService . deleteInstallationService
type deleteInstallationService interface {
	DeleteInstallationAssetCollection() (api.InstallationsServiceOutput, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	GetCurrentInstallationLogs() (api.InstallationsServiceOutput, error)
}

func NewDeleteInstallation(service deleteInstallationService, logger logger) DeleteInstallation {
	return DeleteInstallation{
		service: service,
		logger:  logger,
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

	install, err := ac.service.GetCurrentInstallationLogs()
	if err != nil {
		return fmt.Errorf("installation failed to get logs: %s", err)
	}

	for {
		content, ok := <-install.LogChan
		if ok {
			ac.logger.Println(content)
		} else {
			break
		}
	}

	if err, ok := <-install.ErrorChan; ok {
		if err == api.InstallFailed{
			return errors.New("deleting the installation was unsuccessful")
		}
		return err
	}

	return nil
}

func (ac DeleteInstallation) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command deletes all the products installed on the targeted Ops Manager.",
		ShortDescription: "deletes all the products on the Ops Manager targeted",
	}
}
