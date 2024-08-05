package vmlifecyclecommands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

type initCreateFunc func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.CreateVMService, error)

type CreateVM struct {
	stdout      io.Writer
	stderr      io.Writer
	initService initCreateFunc
	StateFile   string   `long:"state-file" description:"File to output VM identifier info" default:"state.yml"`
	ImageFile   string   `long:"image-file" description:"VM image or yaml map of image locations depending on IaaS" required:"true"`
	Config      string   `long:"config"     description:"The YAML configuration file" required:"true"`
	VarsFile    []string `long:"vars-file"  description:"Load variables from a YAML file for interpolation into config"`
	VarsEnv     []string `long:"vars-env"   env:"OM_VARS_ENV"  description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
}

func NewCreateVMCommand(stdout, stderr io.Writer, initService initCreateFunc) CreateVM {
	return CreateVM{
		stdout:      stdout,
		stderr:      stderr,
		initService: initService,
	}
}

func (c *CreateVM) Execute(args []string) error {
	err := c.checkImageExists()
	if err != nil {
		return err
	}

	config, state, err := loadConfigAndState(c.Config, c.StateFile, false, c.VarsEnv, c.VarsFile)
	if err != nil {
		return err
	}

	vmManagerService, err := c.initService(config, c.ImageFile, state, c.stdout, c.stderr)
	if err != nil {
		return fmt.Errorf("failed to set p-automator: %s", err)
	}

	status, newState, err := vmManagerService.CreateVM()
	switch status {
	case vmmanagers.Success:
		_, _ = c.stdout.Write([]byte("OpsMan VM created successfully\n"))
		return writeStatefile(c.StateFile, newState)
	case vmmanagers.Exist:
		_, _ = c.stdout.Write([]byte("VM already exists, not attempting to create it\n"))
		return writeStatefile(c.StateFile, newState)
	case vmmanagers.StateMismatch:
		return errors.New("VM specified in the statefile does not exist in your IAAS")
	case vmmanagers.Incomplete:
		finalErr := fmt.Errorf("the VM was created, but subsequent configuration failed: %s", err)
		writeErr := writeStatefile(c.StateFile, newState)
		if writeErr != nil {
			return fmt.Errorf("%s\nFailed to write state file: %s", finalErr, writeErr)
		}
		return finalErr
	case vmmanagers.Unknown:
		return fmt.Errorf("unexpected error: %s", err)
	}

	return nil
}

func writeStatefile(filename string, info vmmanagers.StateInfo) error {
	return ioutil.WriteFile(filename, []byte(fmt.Sprintf("iaas: %s\nvm_id: %s", info.IAAS, info.ID)), 0644)
}

func (c *CreateVM) checkImageExists() (err error) {
	_, err = os.Stat(c.ImageFile)
	if err != nil {
		return fmt.Errorf("could not read image file: %s", err)
	}
	return
}
