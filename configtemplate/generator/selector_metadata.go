package generator

import (
	"strings"
)

func SelectorMetadata(optionTemplates []OptionTemplate, selector string) []PropertyMetadata {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.Name
		})
}

// SelectorMetadataBySelectValue - uses the option template SelectValue properties of each OptionTemplate to perform the property medata selection
func SelectorMetadataBySelectValue(optionTemplates []OptionTemplate, selector string) []PropertyMetadata {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.SelectValue
		})
}

func selectorMetadataByFunc(optionTemplates []OptionTemplate, selector string, matchFunc func(optionTemplate OptionTemplate) string) []PropertyMetadata {
	var options []string
	for _, optionTemplate := range optionTemplates {
		match := matchFunc(optionTemplate)
		options = append(options, match)

		if strings.EqualFold(selector, match) {
			return optionTemplate.PropertyMetadata
		}
	}
	return nil
}
