package vmlifecyclecommands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
)

type initCreateFunc func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.CreateVMService, error)

type CreateVM struct {
	stdout      io.Writer
	stderr      io.Writer
	initService initCreateFunc
	StateFile   string   `long:"state-file" description:"File to output VM identifier info" default:"state.yml"`
	ImageFile   string   `long:"image-file" description:"VM image or yaml map of image locations depending on IaaS"`
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
	config, state, err := loadConfigAndState(c.Config, c.StateFile, false, c.VarsEnv, c.VarsFile)
	if err != nil {
		return err
	}

	// Check if image file is required for this IaaS
	err = c.checkImageExists(config)
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
	return os.WriteFile(filename, []byte(fmt.Sprintf("iaas: %s\nvm_id: %s", info.IAAS, info.ID)), 0644)
}

func (c *CreateVM) checkImageExists(config *vmmanagers.OpsmanConfigFilePayload) (err error) {
	// For VCF9, image file is optional if image_name is specified
	if config.OpsmanConfig.VCF9 != nil && config.OpsmanConfig.VCF9.ImageName != "" {
		// Using pre-uploaded image by name, no image file needed
		return nil
	}

	// For all other cases, image file is required
	if c.ImageFile == "" {
		return fmt.Errorf("--image-file is required (or specify image_name in config for VCF9)")
	}

	_, err = os.Stat(c.ImageFile)
	if err != nil {
		return fmt.Errorf("could not read image file: %s", err)
	}
	return
}
