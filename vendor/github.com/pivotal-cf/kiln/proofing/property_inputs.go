package proofing

import yaml "gopkg.in/yaml.v2"

type PropertyInput interface{}

type PropertyInputs []PropertyInput

// TODO: Less ugly.
func (pi *PropertyInputs) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var sniffs []map[string]interface{}
	err := unmarshal(&sniffs)
	if err != nil {
		return err
	}

	var contains = func(m map[string]interface{}, key string) bool {
		_, ok := m[key]
		return ok
	}

	for _, sniff := range sniffs {
		contents, err := yaml.Marshal(sniff)
		if err != nil {
			return err // NOTE: this cannot happen, the YAML has already been unmarshalled
		}

		switch {
		case contains(sniff, "selector_property_inputs"):
			var propertyInput SelectorPropertyInput
			err = yaml.Unmarshal(contents, &propertyInput)
			*pi = append(*pi, propertyInput)
		case contains(sniff, "property_inputs"):
			var propertyInput CollectionPropertyInput
			err = yaml.Unmarshal(contents, &propertyInput)
			*pi = append(*pi, propertyInput)
		default:
			var propertyInput SimplePropertyInput
			err = yaml.Unmarshal(contents, &propertyInput)
			*pi = append(*pi, propertyInput)
		}
		if err != nil {
			return err // NOTE: this cannot happen, the YAML has already been unmarshalled
		}
	}

	return nil
}
