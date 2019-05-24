package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"github.com/pivotal-cf/jhanda"
	"gopkg.in/yaml.v2"
)

type Interpolate struct {
	environFunc func() []string
	logger      logger
	Options     struct {
		ConfigFile        string   `long:"config"       short:"c" description:"path for file to be interpolated"`
		Path              string   `long:"path"                   description:"Extract specified value out of the interpolated file (e.g.: /private_key). The rest of the file will not be printed."`
		VarsEnv           []string `long:"vars-env"               description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		VarsFile          []string `long:"vars-file"    short:"l" description:"Load variables from a YAML file"`
		Vars              []string `long:"var"          short:"v" description:"Load variable from the command line. Format: VAR=VAL"`
		OpsFile           []string `long:"ops-file"     short:"o" description:"YAML operations files"`
		SkipMissingParams bool     `long:"skip-missing" short:"s" description:"Allow skipping missing params"`
	}
}

type interpolateOptions struct {
	templateFile  string
	varsEnvs      []string
	varsFiles     []string
	vars          []string
	opsFiles      []string
	environFunc   func() []string
	expectAllKeys bool
}

func NewInterpolate(environFunc func() []string, logger logger) Interpolate {
	return Interpolate{
		environFunc: environFunc,
		logger:      logger,
	}
}

func (c Interpolate) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse interpolate flags: %s", err)
	}

	input := os.Stdin
	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("error in STDIN: %s", err)
	}

	// Bitwise AND uses stdin's file mode mask against the unix character device to
	// determine if it's pointing to stdin's pipe
	if info.Mode()&os.ModeCharDevice == 0 {
		contents, err := ioutil.ReadAll(input)
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

	} else if len(c.Options.ConfigFile) == 0 {
		return fmt.Errorf("no file or STDIN input provided. Please provide a valid --config file or use a pipe to get STDIN")
	}

	expectAllKeys := true
	if c.Options.SkipMissingParams {
		expectAllKeys = false
	}

	bytes, err := interpolate(interpolateOptions{
		templateFile:  c.Options.ConfigFile,
		varsFiles:     c.Options.VarsFile,
		vars:          c.Options.Vars,
		environFunc:   c.environFunc,
		varsEnvs:      c.Options.VarsEnv,
		opsFiles:      c.Options.OpsFile,
		expectAllKeys: expectAllKeys,
	}, c.Options.Path)
	if err != nil {
		return err
	}

	c.logger.Println(string(bytes))

	return nil
}

func (c Interpolate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "Interpolates variables into a manifest",
		ShortDescription: "Interpolates variables into a manifest",
		Flags:            c.Options,
	}
}

func interpolate(o interpolateOptions, pathStr string) ([]byte, error) {
	contents, err := ioutil.ReadFile(o.templateFile)
	if err != nil {
		return nil, err
	}

	tpl := boshtpl.NewTemplate(contents)
	staticVars := boshtpl.StaticVariables{}
	ops := patch.Ops{}

	for _, varsEnv := range o.varsEnvs {
		for _, envVar := range o.environFunc() {
			pieces := strings.SplitN(envVar, "=", 2)
			if len(pieces) != 2 {
				return []byte{}, errors.New("Expected environment variable to be key-value pair")
			}

			if !strings.HasPrefix(pieces[0], varsEnv+"_") {
				continue
			}

			v := pieces[1]
			var val interface{}
			err = yaml.Unmarshal([]byte(v), &val)
			if err != nil {
				return []byte{}, fmt.Errorf("Could not deserialize YAML from environment variable %q", pieces[0])
			}

			// The environment variable value is treated as YAML, but multi-line strings
			// are line folded, replacing newlines with spaces. If we detect that input value is of
			// type "string" we call yaml.Marshal to ensure characters are escaped properly.
			if fmt.Sprintf("%T", val) == "string" {
				b, _ := yaml.Marshal(v) // err should never occur

				// Don't double quote in the case of an integer that's being used as a string
				// For example, without this regex, a literal string number \"500\"
				// will get unmarshalled as '"500"'
				re := regexp.MustCompile(`^'"\d+"'`)
				if re.Match(b) {
					b = bytes.ReplaceAll(b, []byte(`'`), []byte(""))
				}

				err = yaml.Unmarshal(b, &val)
				if err != nil {
					return []byte{}, fmt.Errorf("Could not deserialize string from environment variable %q", pieces[0])
				}
			}

			staticVars[strings.TrimPrefix(pieces[0], varsEnv+"_")] = val
		}
	}

	for _, path := range o.varsFiles {
		var fileVars boshtpl.StaticVariables
		err = readYAMLFile(path, &fileVars)
		if err != nil {
			return nil, err
		}
		for k, v := range fileVars {
			staticVars[k] = v
		}
	}

	ReadCommandLineVars(o.vars, staticVars)

	for _, path := range o.opsFiles {
		var opDefs []patch.OpDefinition
		err = readYAMLFile(path, &opDefs)
		if err != nil {
			return nil, err
		}
		op, err := patch.NewOpsFromDefinitions(opDefs)
		if err != nil {
			return nil, fmt.Errorf("Building ops (%s)", err.Error())
		}
		ops = append(ops, op)
	}

	evalOpts := boshtpl.EvaluateOpts{
		UnescapedMultiline: true,
		ExpectAllKeys:      o.expectAllKeys,
	}

	path, err := patch.NewPointerFromString(pathStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse path: %s", err)
	}

	if path.IsSet() {
		evalOpts.PostVarSubstitutionOp = patch.FindOp{Path: path}
	}

	bytes, err := tpl.Evaluate(staticVars, ops, evalOpts)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func readYAMLFile(path string, dataType interface{}) error {
	payload, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file (%s): %s", path, err.Error())
	}
	err = yaml.Unmarshal(payload, dataType)
	if err != nil {
		return fmt.Errorf("could not unmarshal file (%s): %s", path, err.Error())
	}

	return nil
}

func ReadCommandLineVars(vars []string, staticVars boshtpl.StaticVariables) {
	for _, singleVar := range vars {
		splitVar := strings.Split(singleVar, "=")

		valInt, err := strconv.Atoi(splitVar[1])
		if err == nil {
			staticVars[splitVar[0]] = valInt
			continue
		}

		valBool, err := strconv.ParseBool(splitVar[1])
		if err == nil {
			staticVars[splitVar[0]] = valBool
			continue
		}

		staticVars[splitVar[0]] = splitVar[1]
	}
}
