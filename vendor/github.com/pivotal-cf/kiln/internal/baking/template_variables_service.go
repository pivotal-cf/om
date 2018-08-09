package baking

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type TemplateVariablesService struct{}

func NewTemplateVariablesService() TemplateVariablesService {
	return TemplateVariablesService{}
}

func (s TemplateVariablesService) FromPathsAndPairs(paths []string, pairs []string) (map[string]interface{}, error) {
	variables := map[string]interface{}{}

	for _, path := range paths {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(content, &variables)
		if err != nil {
			return nil, err
		}
	}

	for _, pair := range pairs {
		parts := strings.Split(pair, "=")

		if len(parts) < 2 {
			return nil, fmt.Errorf("could not parse variable %q: expected variable in \"key=value\" form", pair)
		}

		variables[parts[0]] = parts[1]
	}

	return variables, nil
}
