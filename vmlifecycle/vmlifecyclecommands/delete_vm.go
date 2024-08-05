package vmlifecyclecommands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sort"

	"github.com/pivotal-cf/om/interpolate"
	"github.com/pivotal-cf/om/vmlifecycle/vmmanagers"
	"gopkg.in/yaml.v2"
)

type initDeleteFunc func(config *vmmanagers.OpsmanConfigFilePayload, image string, state vmmanagers.StateInfo, outWriter, errWriter io.Writer) (vmmanagers.DeleteVMService, error)

type DeleteVM struct {
	stdout      io.Writer
	stderr      io.Writer
	initService initDeleteFunc

	StateFile string   `long:"state-file" description:"File containing the VM identifier info" required:"true"`
	Config    string   `long:"config"     description:"The YAML configuration file" required:"true"`
	VarsFile  []string `long:"vars-file"  description:"Load variables from a YAML file for interpolation into config"`
	VarsEnv   []string `long:"vars-env"   env:"OM_VARS_ENV"  description:"load vars from environment variables by specifying a prefix (e.g.: 'MY' to load MY_var=value)"`
}

func NewDeleteVMCommand(stdout, stderr io.Writer, initService initDeleteFunc) DeleteVM {
	return DeleteVM{
		stdout:      stdout,
		stderr:      stderr,
		initService: initService,
	}
}

func (c DeleteVM) Execute(args []string) error {
	config, state, err := loadConfigAndState(c.Config, c.StateFile, true, c.VarsEnv, c.VarsFile)
	if err != nil {
		return err
	}

	vmManagerService, err := c.initService(config, "", state, c.stdout, c.stderr)
	if err != nil {
		return fmt.Errorf("failed to set p-automator: %s", err)
	}

	if state.ID == "" {
		_, _ = c.stdout.Write([]byte("Nothing to do\n"))
		return nil
	}

	err = vmManagerService.DeleteVM()
	if err != nil {
		return fmt.Errorf("delete vm failed, some resources may have not been properly removed: %s", err)
	}

	_, _ = c.stdout.Write([]byte("VM deleted successfully\n"))

	return ioutil.WriteFile(c.StateFile, []byte(fmt.Sprintf(`iaas: %s`, state.IAAS)), 0644)
}

func loadConfigAndState(configFilename string, stateFilename string, useStateFileError bool, varsEnv []string, varsFile []string) (*vmmanagers.OpsmanConfigFilePayload, vmmanagers.StateInfo, error) {
	configContent, err := interpolateConfig(
		configFilename,
		varsFile,
		varsEnv,
	)
	if err != nil {
		return nil, vmmanagers.StateInfo{}, err
	}

	opsmanConfig := &vmmanagers.OpsmanConfigFilePayload{
		OpsmanConfig: vmmanagers.OpsmanConfig{},
		Fields:       nil,
	}
	err = yaml.UnmarshalStrict(configContent, opsmanConfig)
	if err != nil {
		return nil, vmmanagers.StateInfo{}, fmt.Errorf("could not load Opsman Config file (%s): %s", configFilename, err)
	}

	err = GuardAgainstMissingOpsmanConfiguration(configContent, configFilename)
	if err != nil {
		return nil, vmmanagers.StateInfo{}, err
	}

	state := vmmanagers.StateInfo{}
	content, err := ioutil.ReadFile(stateFilename)
	if err != nil && useStateFileError {
		return nil, vmmanagers.StateInfo{}, err
	}

	err = yaml.Unmarshal(content, &state)
	if err != nil {
		return nil, vmmanagers.StateInfo{}, fmt.Errorf("could not load state file (%s): %s", stateFilename, err)
	}

	return opsmanConfig, state, nil
}

func GuardAgainstMissingOpsmanConfiguration(configContent []byte, configFilename string) error {
	var opsmanConfigFound bool
	topLevelkeys := make(map[string]interface{})
	err := yaml.Unmarshal(configContent, topLevelkeys)
	if err != nil {
		return fmt.Errorf("could not load Opsman Config file (%s): %s", configFilename, err)
	}
	fields := reflect.ValueOf(topLevelkeys)
	for _, field := range fields.MapKeys() {
		if field.String() == "opsman-configuration" {
			opsmanConfigFound = true
			break
		}
	}

	if !opsmanConfigFound {
		var keys []string
		for key := range topLevelkeys {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		foundKeysMessage := "Found keys:\n"
		for _, key := range keys {
			foundKeysMessage = foundKeysMessage + fmt.Sprintf("  '%s'\n", key)
		}
		return errors.New("top-level-key 'opsman-configuration' is a required key.\n" +
			"Ensure the correct file is passed, the 'opsman-configuration' key is present, " +
			"and the key is spelled correctly with a dash(-).\n" +
			foundKeysMessage)
	}
	return nil
}

func interpolateConfig(
	config string,
	varsFile []string,
	varsEnv []string,
) ([]byte, error) {
	configContents, err := interpolate.Execute(interpolate.Options{
		TemplateFile:  config,
		VarsFiles:     varsFile,
		EnvironFunc:   os.Environ,
		VarsEnvs:      varsEnv,
		ExpectAllKeys: true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not interpolate config file: %s", err)
	}
	return configContents, err
}
