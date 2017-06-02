package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/flags"
)

type ApplyChanges struct {
	installationsService installationsService
	logger               logger
	logWriter            logWriter
	waitDuration         int
	Options              struct {
		IgnoreWarnings bool `short:"i" long:"ignore-warnings" description:"ignore issues reported by Ops Manager when applying changes"`
	}
}

//go:generate counterfeiter -o ./fakes/installations_service.go --fake-name InstallationsService . installationsService
type installationsService interface {
	Trigger(bool) (api.InstallationsServiceOutput, error)
	Status(id int) (api.InstallationsServiceOutput, error)
	Logs(id int) (api.InstallationsServiceOutput, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
}

//go:generate counterfeiter -o ./fakes/log_writer.go --fake-name LogWriter . logWriter
type logWriter interface {
	Flush(logs string) error
}

func NewApplyChanges(installationsService installationsService, logWriter logWriter, logger logger, waitDuration int) ApplyChanges {
	return ApplyChanges{
		installationsService: installationsService,
		logger:               logger,
		logWriter:            logWriter,
		waitDuration:         waitDuration,
	}
}

func (ac ApplyChanges) Execute(args []string) error {
	_, err := flags.Parse(&ac.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse apply-changes flags: %s", err)
	}

	installation, err := ac.installationsService.RunningInstallation()
	if err != nil {
		return fmt.Errorf("could not check for any already running installation: %s", err)
	}

	if installation == (api.InstallationsServiceOutput{}) {
		ac.logger.Printf("attempting to apply changes to the targeted Ops Manager")
		installation, err = ac.installationsService.Trigger(ac.Options.IgnoreWarnings)
		if err != nil {
			return fmt.Errorf("installation failed to trigger: %s", err)
		}
	} else {
		startedAtFormatted := installation.StartedAt.Format(time.UnixDate)
		ac.logger.Printf("found already running installation...re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)
	}

	for {
		current, err := ac.installationsService.Status(installation.ID)
		if err != nil {
			return fmt.Errorf("installation failed to get status: %s", err)
		}

		install, err := ac.installationsService.Logs(installation.ID)
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
			return errors.New("installation was unsuccessful")
		}

		time.Sleep(time.Duration(ac.waitDuration) * time.Second)
	}
}

func (ac ApplyChanges) Usage() Usage {
	return Usage{
		Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
		ShortDescription: "triggers an install on the Ops Manager targeted",
		Flags:            ac.Options,
	}
}
