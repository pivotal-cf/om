package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ApplyChanges struct {
	service      applyChangesService
	logger       logger
	logWriter    logWriter
	waitDuration int
	Options      struct {
		IgnoreWarnings     bool `short:"i"   long:"ignore-warnings"      description:"ignore issues reported by Ops Manager when applying changes"`
		SkipDeployProducts bool `short:"sdp" long:"skip-deploy-products" description:"skip deploying products when applying changes - just update the director"`
	}
}

//go:generate counterfeiter -o ./fakes/apply_changes_service.go --fake-name ApplyChangesService . applyChangesService
type applyChangesService interface {
	CreateInstallation(bool, bool) (api.InstallationsServiceOutput, error)
	GetInstallation(id int) (api.InstallationsServiceOutput, error)
	GetInstallationLogs(id int) (api.InstallationsServiceOutput, error)
	RunningInstallation() (api.InstallationsServiceOutput, error)
	ListInstallations() ([]api.InstallationsServiceOutput, error)
}

//go:generate counterfeiter -o ./fakes/log_writer.go --fake-name LogWriter . logWriter
type logWriter interface {
	Flush(logs string) error
}

func NewApplyChanges(service applyChangesService, logWriter logWriter, logger logger, waitDuration int) ApplyChanges {
	return ApplyChanges{
		service:      service,
		logger:       logger,
		logWriter:    logWriter,
		waitDuration: waitDuration,
	}
}

func (ac ApplyChanges) Execute(args []string) error {
	if _, err := jhanda.Parse(&ac.Options, args); err != nil {
		return fmt.Errorf("could not parse apply-changes flags: %s", err)
	}

	installation, err := ac.service.RunningInstallation()
	if err != nil {
		return fmt.Errorf("could not check for any already running installation: %s", err)
	}

	if installation == (api.InstallationsServiceOutput{}) {
		ac.logger.Printf("attempting to apply changes to the targeted Ops Manager")
		deployProducts := !ac.Options.SkipDeployProducts
		installation, err = ac.service.CreateInstallation(ac.Options.IgnoreWarnings, deployProducts)
		if err != nil {
			return fmt.Errorf("installation failed to trigger: %s", err)
		}
	} else {
		startedAtFormatted := installation.StartedAt.Format(time.UnixDate)
		ac.logger.Printf("found already running installation...re-attaching (Installation ID: %d, Started: %s)", installation.ID, startedAtFormatted)
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
			return errors.New("installation was unsuccessful")
		}

		time.Sleep(time.Duration(ac.waitDuration) * time.Second)
	}
}

func (ac ApplyChanges) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
		ShortDescription: "triggers an install on the Ops Manager targeted",
		Flags:            ac.Options,
	}
}
