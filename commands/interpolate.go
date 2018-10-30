package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
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
		ConfigFile string   `long:"config"    short:"c" required:"true" description:"path for file to be interpolated"`
		Path       string   `long:"path"                                description:"Extract specified value out of the interpolated file (e.g.: /private_key). The rest of the file will not be printed."`
		VarsEnv    []string `long:"vars-env"                            description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
		VarsFile   []string `long:"vars-file" short:"l"                 description:"Load variables from a YAML file"`
		OpsFile    []string `long:"ops-file"  short:"o"                 description:"YAML operations files"`
	}
}

type interpolateOptions struct {
	templateFile string
	varsEnvs     []string
	varsFiles    []string
	opsFiles     []string
	environFunc  func() []string
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

	bytes, err := interpolate(interpolateOptions{
		templateFile: c.Options.ConfigFile,
		varsFiles:    c.Options.VarsFile,
		environFunc:  c.environFunc,
		varsEnvs:     c.Options.VarsEnv,
		opsFiles:     c.Options.OpsFile,
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
		ExpectAllKeys:      true,
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
