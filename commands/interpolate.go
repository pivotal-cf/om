package commands

import (
	"fmt"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/pivotal-cf/jhanda"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Interpolate struct {
	logger  logger
	Options struct {
		ConfigFile string   `long:"config" short:"c" required:"true" description:"path for file to be interpolated"`
		OutputFile string   `long:"output-file" short:"o" description:"output file for interpolated YAML"`
		VarsFile   []string `long:"vars-file" short:"l" description:"Load variables from a YAML file"`
	}
}

func NewInterpolate(logger logger) Interpolate {
	return Interpolate{
		logger: logger,
	}
}

func (c Interpolate) Execute(args []string) error {
	if _, err := jhanda.Parse(&c.Options, args); err != nil {
		return fmt.Errorf("could not parse interpolate flags: %s", err)
	}

	bytes, err := interpolate(c.Options.ConfigFile, c.Options.VarsFile)
	if err != nil {
		return err
	}

	if c.Options.OutputFile != "" {
		return ioutil.WriteFile(c.Options.OutputFile, bytes, os.ModePerm)
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

func interpolate(templateFile string, varsFiles []string) ([]byte, error) {
	contents, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}

	tpl := boshtpl.NewTemplate(contents)
	vars := []boshtpl.Variables{}

	for _, path := range varsFiles {
		payload, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read template variables file (%s): %s", path, err.Error())
		}
		var staticVars boshtpl.StaticVariables
		err = yaml.Unmarshal(payload, &staticVars)
		if err != nil {
			return nil, fmt.Errorf("could not unmarhsal template variables file (%s): %s", path, err.Error())
		}
		vars = append(vars, staticVars)
	}
	evalOpts := boshtpl.EvaluateOpts{
		UnescapedMultiline: true,
		ExpectAllKeys:      true,
	}

	bytes, err := tpl.Evaluate(boshtpl.NewMultiVars(vars), nil, evalOpts)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
