package cmd

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"strconv"

	"github.com/pivotal-cf/om/interpolate"
	"gopkg.in/yaml.v2"
)

// Load the config file, (optionally) load the vars file, vars env as well
// To use this function, `Config` field must be defined in the command struct being passed in.
// To load vars, VarsFile and/or VarsEnv must exist in the command struct being passed in.
// If VarsEnv is used, envFunc must be defined instead of nil
func loadConfigFile(args []string, envFunc func() []string) ([]string, error) {
	for _, cmdConfigBypassList := range []string{
		"interpolate",
		"configure-opsman",
		"configure-product",
		"configure-director",
	} {
		if cmdConfigBypassList == args[0] {
			return args, nil
		}
	}
	
	var err error
	var config struct {
		ConfigFile string   `long:"config"                     short:"c"`
		VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV"`
		VarsFile   []string `long:"vars-file"                  short:"l"`
		Vars       []string `long:"var"                        short:"v"`
	}

	parser := flags.NewParser(&config, flags.IgnoreUnknown)
	args, err = parser.ParseArgs(args)
	configFile := config.ConfigFile
	if configFile == "" {
		return args, err
	}

	var (
		cmdVars []string
		options map[string]interface{}
	)

	contents, err := interpolate.Execute(interpolate.Options{
		TemplateFile:  configFile,
		VarsEnvs:      config.VarsEnv,
		VarsFiles:     config.VarsFile,
		Vars:          cmdVars,
		EnvironFunc:   envFunc,
		OpsFiles:      nil,
		ExpectAllKeys: true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not load the config file: %s", err)
	}

	err = yaml.Unmarshal(contents, &options)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %s", configFile, err)
	}

	var fileArgs []string
	for key, value := range options {
		switch convertedValue := value.(type) {
		case []interface{}:
			for _, v := range convertedValue {
				fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", key, v))
			}
		case bool:
			fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", key, strconv.FormatBool(convertedValue)))
		default:
			fileArgs = append(fileArgs, fmt.Sprintf("--%s=%s", key, value))
		}

	}
	fileArgs = append(args, fileArgs...)
	return fileArgs, err
}
