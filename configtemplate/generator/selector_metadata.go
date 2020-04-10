package generator

import (
	"strings"
)

func SelectorOptionsBlueprints(optionTemplates []OptionTemplate, selector string) []PropertyBlueprint {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.Name
		})
}

// SelectorBlueprintsBySelectValue - uses the option template SelectValue properties of each OptionTemplate to perform the property medata selection
func SelectorBlueprintsBySelectValue(optionTemplates []OptionTemplate, selector string) []PropertyBlueprint {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.SelectValue
		})
}

func selectorMetadataByFunc(optionTemplates []OptionTemplate, selector string, matchFunc func(optionTemplate OptionTemplate) string) []PropertyBlueprint {
	for _, optionTemplate := range optionTemplates {
		match := matchFunc(optionTemplate)

		if strings.EqualFold(selector, match) {
			return optionTemplate.PropertyBlueprints
		}
	}
	return nil
}
