package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type DeleteInstallation struct {
	service      deleteInstallationService
	logger       logger
	logWriter    logWriter
	stdin        io.Reader
	waitDuration time.Duration
	Options      struct {
		Force bool `long:"force" short:"f" description:"used to avoid interactive prompt acknowledging deletion"`
	}
}

//go:generate counterfeiter -o ./fakes/delete_installation_service.go --fake-name DeleteInstallationService . deleteInstallationService
type deleteInstallationService interface {
	DeleteInstallationAssetCollection() (api.InstallationsServiceOutput, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	GetInstallation(id int) (api.InstallationsServiceOutput, error)
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
}

func NewDeleteInstallation(service deleteInstallationService, logWriter logWriter, logger logger, stdin io.Reader, waitDuration time.Duration) DeleteInstallation {
	return DeleteInstallation{
		service:      service,
		logger:       logger,
		logWriter:    logWriter,
		stdin:        stdin,
		waitDuration: waitDuration,
	}
}

func (ac DeleteInstallation) Execute(args []string) error {
	if _, err := jhanda.Parse(&ac.Options, args); err != nil {
		return fmt.Errorf("could not parse delete-installation flags: %s", err)
	}

	if !ac.Options.Force {
		for {
			scanner := bufio.NewScanner(ac.stdin)
			ac.logger.Printf("Do you really want to delete the installation? [yes/no]: ")

			scanner.Scan()
			text := scanner.Text()

			if text == "yes" {
				ac.logger.Printf("Ok. Are you sure? [yes/no]: ")
				scanner.Scan()
				text = scanner.Text()

				if text == "yes" {
					ac.logger.Printf("Ok, deleting installation.")
					break
				}
			}

			ac.logger.Printf("Ok, nothing was deleted.")
			return nil
		}
	}

	//we aren't chechking this error for now
	installation, _ := ac.service.RunningInstallation()

	if installation == (api.InstallationsServiceOutput{}) {
		var err error

		ac.logger.Printf("attempting to delete the installation on the targeted Ops Manager")

		installation, err = ac.service.DeleteInstallationAssetCollection()
		if err != nil {
			return fmt.Errorf("failed to delete installation: %s", err)
		}

		if installation == (api.InstallationsServiceOutput{}) {
			ac.logger.Printf("no installation to delete")
			return nil
		}
	}

	ac.logger.Printf("found already running deletion...attempting to re-attach")

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
		Flags:            ac.Options,
	}
}
