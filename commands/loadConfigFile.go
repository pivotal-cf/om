package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"gopkg.in/yaml.v2"
	"reflect"
)

// Load the config file, (optionally) load the vars file, vars env as well
// To use this function, `Config` field must be defined in the command struct being passed in.
// To load vars, VarsFile and/or VarsEnv must exist in the command struct being passed in.
// If VarsEnv is used, envFunc must be defined instead of nil
func loadConfigFile(args []string, command interface{}, envFunc func() []string) error {
	_, err := jhanda.Parse(command, args)
	commandValue := reflect.ValueOf(command).Elem()
	configFile := commandValue.FieldByName("ConfigFile").String()
	if configFile == "" {
		return err
	}

	varsFileField := commandValue.FieldByName("VarsFile")
	varsEnvField := commandValue.FieldByName("VarsEnv")
	cmdVarsField := commandValue.FieldByName("Vars")

	var (
		varsField []string
		varsEnv   []string
		cmdVars   []string
		ok        bool
		options   map[string]interface{}
		contents  []byte
	)

	if varsFileField.IsValid() {
		if varsField, ok = varsFileField.Interface().([]string); !ok {
			return fmt.Errorf("expect VarsFile field to be a `[]string`, found %s", varsEnvField.Type())
		}
	}

	if cmdVarsField.IsValid() {
		if cmdVars, ok = cmdVarsField.Interface().([]string); !ok {
			return fmt.Errorf("expect Vars field to be a `[]string`, found %s", cmdVarsField.Type())
		}
	}

	if varsEnvField.IsValid() {
		if varsEnv, ok = varsEnvField.Interface().([]string); !ok {
			return fmt.Errorf("expect VarsEnv field to be a `[]string`, found %s", varsEnvField.Type())
		}
	}

	contents, err = interpolate(interpolateOptions{
		templateFile:  configFile,
		varsEnvs:      varsEnv,
		varsFiles:     varsField,
		vars:          cmdVars,
		environFunc:   envFunc,
		opsFiles:      nil,
		expectAllKeys: true,
	}, "")
	if err != nil {
		return fmt.Errorf("could not load the config file: %s", err)
	}

	err = yaml.Unmarshal(contents, &options)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file %s: %s", configFile, err)
	}

	var fileArgs []string
	for key, value := range options {
		switch convertedValue := value.(type) {
		case []interface{}:
			for _, v := range convertedValue {
				fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", key, v))
			}
		default:
			fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", key, value))
		}

	}
	fileArgs = append(fileArgs, args...)
	_, err = jhanda.Parse(command, fileArgs)
	return err
}
