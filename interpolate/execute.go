package interpolate

import (
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

	// the following was taken from bosh cli
	// https://github.com/cloudfoundry/bosh-cli/blob/9c1c210c83673a780e3787a91f444541755e6585/cmd/opts/var_flags.go
	// we cannot use it directly because of the use of `jhanda`
	staticVars := template.StaticVariables{}

	for _, v := range o.VarsEnvs {
		varsEnvArg := &template.VarsEnvArg{EnvironFunc: o.EnvironFunc}
		err := varsEnvArg.UnmarshalFlag(v)
		if err != nil {
			return nil, err
		}

		for k, v := range varsEnvArg.Vars {
			staticVars[k] = v
		}
	}

	for _, v := range o.VarsFiles {
		varFilesArg := &template.VarsFileArg{}
		err := varFilesArg.UnmarshalFlag(v)
		if err != nil {
			return nil, err
		}

		for k, v := range varFilesArg.Vars {
			staticVars[k] = v
		}
	}

	for _, v := range o.Vars {
		varArg := &template.VarKV{}
		err := varArg.UnmarshalFlag(v)
		if err != nil {
			return nil, err
		}

		staticVars[varArg.Name] = varArg.Value
	}

	if err != nil {
		return nil, err
	}

	ops := patch.Ops{}
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
