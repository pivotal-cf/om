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

	var (
		varsField []string
		varsEnv   []string
		ok        bool
		options   map[string]string
		contents  []byte
	)

	if varsFileField.IsValid() {
		if varsField, ok = varsFileField.Interface().([]string); !ok {
			return fmt.Errorf("expect VarsFile field to be a `[]string`, found %s", varsEnvField.Type())
		}
	}

	if varsEnvField.IsValid() {
		if varsEnv, ok = varsEnvField.Interface().([]string); !ok {
			return fmt.Errorf("expect VarsEnv field to be a `[]string`, found %s", varsEnvField.Type())
		}
	}

	contents, err = interpolate(interpolateOptions{
		templateFile: configFile,
		varsEnvs:     varsEnv,
		varsFiles:    varsField,
		environFunc:  envFunc,
		opsFiles:     nil,
	}, "")
	if err != nil {
		return fmt.Errorf("could not load the config file: %s", err)
	}

	err = yaml.Unmarshal(contents, &options)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file %s: %s", configFile, err)
	}

	var fileArgs []string
	for k, v := range options {
		fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", k, v))
	}
	fileArgs = append(fileArgs, args...)
	_, err = jhanda.Parse(command, fileArgs)
	return err
}
