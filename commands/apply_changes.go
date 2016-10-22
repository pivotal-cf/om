package commands

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/common"
)

type ApplyChanges struct {
	installationsService installationsService
	logger               common.Logger
	logWriter            logWriter
	waitDuration         int
	Options              struct{}
}

//go:generate counterfeiter -o ./fakes/installations_service.go --fake-name InstallationsService . installationsService
type installationsService interface {
	Trigger() (api.InstallationsServiceOutput, error)
	Status(id int) (api.InstallationsServiceOutput, error)
	Logs(id int) (api.InstallationsServiceOutput, error)
}

//go:generate counterfeiter -o ./fakes/log_writer.go --fake-name LogWriter . logWriter
type logWriter interface {
	Flush(logs string) error
}

func NewApplyChanges(installationsService installationsService, logWriter logWriter, logger common.Logger, waitDuration int) ApplyChanges {
	return ApplyChanges{
		installationsService: installationsService,
		logger:               logger,
		logWriter:            logWriter,
		waitDuration:         waitDuration,
	}
}

func (ac ApplyChanges) Execute(args []string) error {
	ac.logger.Printf("attempting to apply changes to the targeted Ops Manager")

	installation, err := ac.installationsService.Trigger()
	if err != nil {
		return fmt.Errorf("installation failed to trigger: %s", err)
	}

	for {
		current, err := ac.installationsService.Status(installation.ID)
		ok, err := checkErr(err)
		if err != nil {
			return fmt.Errorf("installation failed to get status: %s", err)
		}

		if ok {
			install, err := ac.installationsService.Logs(installation.ID)
			ok, err = checkErr(err)
			if err != nil {
				return fmt.Errorf("installation failed to get logs: %s", err)
			}

			if ok {
				err = ac.logWriter.Flush(install.Logs)
				if err != nil {
					return fmt.Errorf("installation failed to flush logs: %s", err)
				}

				if current.Status == api.StatusSucceeded {
					return nil
				} else if current.Status == api.StatusFailed {
					return errors.New("installation was unsuccessful")
				}
			}
		}

		time.Sleep(time.Duration(ac.waitDuration) * time.Second)
	}
}

func checkErr(err error) (bool, error) {
	if err != nil {
		if ne, ok := err.(net.Error); ok {
			if ne.Temporary() {
				return false, nil
			}
		}

		if err == io.EOF {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (ac ApplyChanges) Usage() Usage {
	return Usage{
		Description:      "This authenticated command kicks off an install of any staged changes on the Ops Manager.",
		ShortDescription: "triggers an install on the Ops Manager targeted",
		Flags:            ac.Options,
	}
}
