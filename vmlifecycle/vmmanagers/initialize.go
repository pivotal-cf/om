package vmmanagers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pivotal-cf/om/vmlifecycle/runner"
	"reflect"
	"time"
)

type vmManager interface {
	CreateVM() (status Status, state StateInfo, err error)
	DeleteVM() error
}

//go:generate counterfeiter -o ./fakes/delete_vm.go --fake-name DeleteVMService . DeleteVMService
type DeleteVMService interface {
	DeleteVM() error
}

//go:generate counterfeiter -o ./fakes/create_vm.go --fake-name CreateVMService . CreateVMService
type CreateVMService interface {
	CreateVM() (Status, StateInfo, error)
}

func NewDeleteVMManager(config *OpsmanConfigFilePayload, image string, state StateInfo, outWriter, errWriter io.Writer) (DeleteVMService, error) {
	return initializeVMManager(config, image, state, outWriter, errWriter)
}

func NewCreateVMManager(config *OpsmanConfigFilePayload, image string, state StateInfo, outWriter, errWriter io.Writer) (CreateVMService, error) {
	return initializeVMManager(config, image, state, outWriter, errWriter)
}

func initializeVMManager(config *OpsmanConfigFilePayload, image string, state StateInfo, outWriter, errWriter io.Writer) (vmManager, error) {
	if err := ValidateOpsManConfig(config); err != nil {
		return nil, err
	}

	if config.OpsmanConfig.Vsphere != nil {
		_, _ = outWriter.Write([]byte(fmt.Sprintln("Using vSphere...")))
		govcCLI, err := runner.NewRunner("govc", outWriter, errWriter)
		if err != nil {
			return nil, err
		}

		return NewVsphereVMManager(
			config,
			image,
			state,
			govcCLI,
		), nil
	}

	if config.OpsmanConfig.GCP != nil {
		_, _ = outWriter.Write([]byte(fmt.Sprintln("Using gcp...")))
		gcpCLI, err := runner.NewRunner("gcloud", outWriter, errWriter)
		if err != nil {
			return nil, err
		}

		return NewGcpVMManager(
			config,
			image,
			state,
			gcpCLI,
		), nil
	}

	if config.OpsmanConfig.AWS != nil {
		_, _ = outWriter.Write([]byte(fmt.Sprintln("Using aws...")))
		awsCLI, err := runner.NewRunner("aws", outWriter, errWriter)
		if err != nil {
			return nil, err
		}

		return NewAWSVMManager(
			os.Stdout,
			os.Stderr,
			config,
			image,
			state,
			awsCLI,
			5*time.Second,
		), nil
	}

	if config.OpsmanConfig.Azure != nil {
		_, _ = outWriter.Write([]byte(fmt.Sprintln("Using azure...")))
		azureCLI, err := runner.NewRunner("az", outWriter, errWriter)
		if err != nil {
			return nil, err
		}

		return NewAzureVMManager(
			os.Stdout,
			os.Stderr,
			config,
			image,
			state,
			azureCLI,
			5*time.Second,
		), nil
	}

	if config.OpsmanConfig.Openstack != nil {
		_, _ = outWriter.Write([]byte(fmt.Sprintln("Using openstack...")))
		openstackCLI, err := runner.NewRunner("openstack", outWriter, errWriter)
		if err != nil {
			return nil, err
		}

		return NewOpenstackVMManager(
			config,
			image,
			state,
			openstackCLI,
		), nil
	}

	return nil, errors.New("unexpected error")
}

func ValidateOpsManConfig(config *OpsmanConfigFilePayload) error {
	if len(config.OpsmanConfig.Unknown) > 0 {
		var unknownIaas []string
		for key := range config.OpsmanConfig.Unknown {
			unknownIaas = append(unknownIaas, key)
		}

		return fmt.Errorf("unknown iaas: %v, please refer to documentation: %s", strings.Join(unknownIaas, ", "), "github.com/pivotal/platform-automation/docstest")
	}

	fields := reflect.ValueOf(&config.OpsmanConfig).Elem()

	count := 0
	for i := 0; i < fields.NumField(); i++ {
		f := fields.Field(i)
		if !f.IsNil() {
			count += 1
		}
	}

	if count > 1 {
		return errors.New("more than one iaas matched, only one in config allowed")
	}
	if count == 0 {
		return errors.New("no iaas configuration provided, please refer to documentation")
	}

	return nil
}
