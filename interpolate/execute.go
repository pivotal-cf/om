package interpolate

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

type Options struct {
	TemplateFile  string
	VarsEnvs      []string
	VarsFiles     []string
	Vars          []string
	OpsFiles      []string
	EnvironFunc   func() []string
	ExpectAllKeys bool
	Path          string
}

func Execute(o Options) ([]byte, error) {
	contents, err := ioutil.ReadFile(o.TemplateFile)
	if err != nil {
		return nil, fmt.Errorf("could not read file (%s): %s", o.TemplateFile, err.Error())
	}

	tpl := template.NewTemplate(contents)
	staticVars := template.StaticVariables{}
	ops := patch.Ops{}

	for _, varsEnv := range o.VarsEnvs {
		for _, envVar := range o.EnvironFunc() {

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

	for _, path := range o.VarsFiles {
		var fileVars template.StaticVariables
		err = readYAMLFile(path, &fileVars)
		if err != nil {
			return nil, err
		}
		for k, v := range fileVars {
			staticVars[k] = v
		}
	}

	readCommandLineVars(o.Vars, staticVars)

	for _, path := range o.OpsFiles {
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

	evalOpts := template.EvaluateOpts{
		UnescapedMultiline: true,
		ExpectAllKeys:      o.ExpectAllKeys,
	}

	path, err := patch.NewPointerFromString(o.Path)
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

func readCommandLineVars(vars []string, staticVars template.StaticVariables) {
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
