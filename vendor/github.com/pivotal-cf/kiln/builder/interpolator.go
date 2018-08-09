package builder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	yamlConverter "github.com/ghodss/yaml"
	yaml "gopkg.in/yaml.v2"
)

type Interpolator struct{}

type InterpolateInput struct {
	Version            string
	BOSHVariables      map[string]interface{}
	Variables          map[string]interface{}
	ReleaseManifests   map[string]interface{}
	StemcellManifest   interface{}
	FormTypes          map[string]interface{}
	IconImage          string
	InstanceGroups     map[string]interface{}
	Jobs               map[string]interface{}
	PropertyBlueprints map[string]interface{}
	RuntimeConfigs     map[string]interface{}
}

func NewInterpolator() Interpolator {
	return Interpolator{}
}

func (i Interpolator) Interpolate(input InterpolateInput, templateYAML []byte) ([]byte, error) {
	interpolatedYAML, err := i.interpolate(input, templateYAML)
	if err != nil {
		return nil, err
	}

	prettyMetadata, err := i.prettyPrint(interpolatedYAML)
	if err != nil {
		return nil, err // un-tested
	}

	return prettyMetadata, nil
}

func (i Interpolator) interpolate(input InterpolateInput, templateYAML []byte) ([]byte, error) {
	templateHelpers := template.FuncMap{
		"bosh_variable": func(key string) (string, error) {
			if input.BOSHVariables == nil {
				return "", errors.New("--bosh-variables-directory must be specified")
			}
			val, ok := input.BOSHVariables[key]
			if !ok {
				return "", fmt.Errorf("could not find bosh variable with key '%s'", key)
			}
			return i.interpolateValueIntoYAML(input, val)
		},
		"form": func(key string) (string, error) {
			if input.FormTypes == nil {
				return "", errors.New("--forms-directory must be specified")
			}
			val, ok := input.FormTypes[key]
			if !ok {
				return "", fmt.Errorf("could not find form with key '%s'", key)
			}

			return i.interpolateValueIntoYAML(input, val)
		},
		"property": func(name string) (string, error) {
			if input.PropertyBlueprints == nil {
				return "", errors.New("--properties-directory must be specified")
			}
			val, ok := input.PropertyBlueprints[name]
			if !ok {
				return "", fmt.Errorf("could not find property blueprint with name '%s'", name)
			}
			return i.interpolateValueIntoYAML(input, val)
		},
		"release": func(name string) (string, error) {
			if input.ReleaseManifests == nil {
				return "", errors.New("--releases-directory must be specified")
			}
			val, ok := input.ReleaseManifests[name]
			if !ok {
				return "", fmt.Errorf("could not find release with name '%s'", name)
			}

			return i.interpolateValueIntoYAML(input, val)
		},
		"stemcell": func() (string, error) {
			if input.StemcellManifest == nil {
				return "", errors.New("--stemcell-tarball must be specified")
			}
			return i.interpolateValueIntoYAML(input, input.StemcellManifest)
		},
		"version": func() (string, error) {
			if input.Version == "" {
				return "", errors.New("--version must be specified")
			}
			return i.interpolateValueIntoYAML(input, input.Version)
		},
		"variable": func(key string) (string, error) {
			if input.Variables == nil {
				return "", errors.New("--variable or --variables-file must be specified")
			}
			val, ok := input.Variables[key]
			if !ok {
				return "", fmt.Errorf("could not find variable with key '%s'", key)
			}
			return i.interpolateValueIntoYAML(input, val)
		},
		"icon": func() (string, error) {
			if input.IconImage == "" {
				return "", fmt.Errorf("--icon must be specified")
			}
			return input.IconImage, nil
		},
		"instance_group": func(name string) (string, error) {
			if input.InstanceGroups == nil {
				return "", errors.New("--instance-groups-directory must be specified")
			}
			val, ok := input.InstanceGroups[name]
			if !ok {
				return "", fmt.Errorf("could not find instance_group with name '%s'", name)
			}

			return i.interpolateValueIntoYAML(input, val)
		},
		"job": func(name string) (string, error) {
			if input.Jobs == nil {
				return "", errors.New("--jobs-directory must be specified")
			}
			val, ok := input.Jobs[name]
			if !ok {
				return "", fmt.Errorf("could not find job with name '%s'", name)
			}

			return i.interpolateValueIntoYAML(input, val)
		},
		"runtime_config": func(name string) (string, error) {
			if input.RuntimeConfigs == nil {
				return "", errors.New("--runtime-configs-directory must be specified")
			}
			val, ok := input.RuntimeConfigs[name]
			if !ok {
				return "", fmt.Errorf("could not find runtime_config with name '%s'", name)
			}

			return i.interpolateValueIntoYAML(input, val)
		},
		"select": func(field, input string) (string, error) {
			object := map[string]interface{}{}

			err := json.Unmarshal([]byte(input), &object)
			if err != nil {
				return "", fmt.Errorf("could not JSON unmarshal %q: %s", input, err)
			}

			value, ok := object[field]
			if !ok {
				return "", fmt.Errorf("could not select %q, key does not exist", field)
			}

			output, err := json.Marshal(value)
			if err != nil {
				return "", fmt.Errorf("could not JSON marshal %q: %s", input, err) // NOTE: this cannot happen because value was unmarshalled from JSON
			}

			return string(output), nil
		},
	}

	t, err := template.New("metadata").
		Delims("$(", ")").
		Funcs(templateHelpers).
		Parse(string(templateYAML))

	if err != nil {
		return nil, fmt.Errorf("template parsing failed: %s", err)
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, input.Variables)
	if err != nil {
		return nil, fmt.Errorf("template execution failed: %s", err)
	}

	return buffer.Bytes(), nil
}

func (i Interpolator) interpolateValueIntoYAML(input InterpolateInput, val interface{}) (string, error) {
	initialYAML, err := yaml.Marshal(val)
	if err != nil {
		return "", err // should never happen
	}

	interpolatedYAML, err := i.interpolate(input, initialYAML)
	if err != nil {
		return "", fmt.Errorf("unable to interpolate value: %s", err)
	}

	inlinedYAML, err := i.yamlMarshalOneLine(interpolatedYAML)
	if err != nil {
		return "", err // un-tested
	}

	return string(inlinedYAML), nil
}

// Workaround to avoid YAML indentation being incorrect when value is interpolated into the metadata
func (i Interpolator) yamlMarshalOneLine(yamlContents []byte) ([]byte, error) {
	return yamlConverter.YAMLToJSON(yamlContents)
}

func (i Interpolator) prettifyRuntimeConfig(interpolatedYAML string) (string, error) {
	var runtimeConfig map[string]interface{}
	err := yaml.Unmarshal([]byte(interpolatedYAML), &runtimeConfig)
	if err != nil {
		return "", err
	}

	if _, ok := runtimeConfig["runtime_config"]; !ok {
		return interpolatedYAML, err
	}

	prettyRuntimeConfig, err := i.prettyPrint([]byte(runtimeConfig["runtime_config"].(string)))
	if err != nil {
		return "", err
	}

	runtimeConfig["runtime_config"] = string(prettyRuntimeConfig)

	prettyInterpolatedYAML, err := yaml.Marshal(runtimeConfig)
	if err != nil {
		return "", err // should never happen
	}

	inlinedYAML, err := i.yamlMarshalOneLine(prettyInterpolatedYAML)
	if err != nil {
		return "", err // un-tested
	}

	return string(inlinedYAML), nil
}

func (i Interpolator) prettyPrint(inputYAML []byte) ([]byte, error) {
	var data interface{}
	err := yaml.Unmarshal(inputYAML, &data)
	if err != nil {
		return []byte{}, err // should never happen
	}

	return yaml.Marshal(data)
}
