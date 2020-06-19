package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/interpolate"
	"io/ioutil"
	"os"
)

type interpolateOptions struct {
	ConfigFile        string   `long:"config"       short:"c"     description:"path for file to be interpolated"`
	VarsEnv           []string `long:"vars-env" env:"OM_VARS_ENV" description:"load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)"`
	VarsFile          []string `long:"vars-file"    short:"l"     description:"load variables from a YAML file"`
	Vars              []string `long:"var"          short:"v"     description:"load variable from the command line. Format: VAR=VAL"`
}

type interpolateConfigFileOptions struct {
	ConfigFile string   `long:"config"                     short:"c" description:"path to yml file for configuration (keys must match the following command line flags)"`
	VarsEnv    []string `long:"vars-env" env:"OM_VARS_ENV"           description:"load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)"`
	VarsFile   []string `long:"vars-file"                  short:"l" description:"load variables from a YAML file"`
	Vars       []string `long:"var"                        short:"v" description:"load variable from the command line. Format: VAR=VAL"`
}

type Interpolate struct {
	environFunc func() []string
	logger      logger
	input       *os.File
	Options     struct {
		interpolateOptions
		Path              string   `long:"path"                       description:"extract specified value out of the interpolated file (e.g.: /private_key). The rest of the file will not be printed."`
		OpsFile           []string `long:"ops-file"     short:"o"     description:"YAML operations files"`
		SkipMissingParams bool     `long:"skip-missing" short:"s"     description:"allow skipping missing params"`
	}
}

func NewInterpolate(environFunc func() []string, logger logger, input *os.File) Interpolate {
	return Interpolate{
		environFunc: environFunc,
		logger:      logger,
		input:       input,
	}
}

func (c Interpolate) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse interpolate flags: %s", err)
	}

	info, err := c.input.Stat()
	if err != nil {
		return fmt.Errorf("error in STDIN: %s", err)
	}

	// Bitwise AND uses stdin's file mode mask against the unix character device to
	// determine if it's pointing to stdin's pipe
	if info.Mode()&os.ModeCharDevice == 0 && (len(c.Options.ConfigFile) == 0 || c.Options.ConfigFile == "-") {
		contents, err := ioutil.ReadAll(c.input)
		if err != nil {
			return fmt.Errorf("error reading STDIN: %s", err)
		}

		tempFile, err := ioutil.TempFile("", "yml")
		if err != nil {
			return fmt.Errorf("error generating temp file for STDIN: %s", err)
		}

		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write(contents)
		if err != nil {
			return fmt.Errorf("error writing temp file for STDIN: %s", err)
		}

		c.Options.ConfigFile = tempFile.Name()

	} else if len(c.Options.ConfigFile) == 0 || c.Options.ConfigFile == "-" {
		return fmt.Errorf("no file or STDIN input provided. Please provide a valid --config file or use a pipe to get STDIN")
	}

	expectAllKeys := true
	if c.Options.SkipMissingParams {
		expectAllKeys = false
	}

	bytes, err := interpolate.Execute(interpolate.Options{
		TemplateFile:  c.Options.ConfigFile,
		VarsFiles:     c.Options.VarsFile,
		Vars:          c.Options.Vars,
		EnvironFunc:   c.environFunc,
		VarsEnvs:      c.Options.VarsEnv,
		OpsFiles:      c.Options.OpsFile,
		ExpectAllKeys: expectAllKeys,
		Path:          c.Options.Path,
	})
	if err != nil {
		return err
	}

	c.logger.Print(string(bytes))

	return nil
}

func (c Interpolate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "interpolates variables into a manifest",
		ShortDescription: "interpolates variables into a manifest",
		Flags:            c.Options,
	}
}
