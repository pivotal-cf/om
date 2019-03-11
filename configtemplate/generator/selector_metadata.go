package generator

import (
	"fmt"
	"strings"
)

func SelectorMetadata(optionTemplates []OptionTemplate, selector string) ([]PropertyMetadata, error) {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.Name
		})
}

// SelectorMetadataBySelectValue - uses the option template SelectValue properties of each OptionTemplate to perform the property medata selection
func SelectorMetadataBySelectValue(optionTemplates []OptionTemplate, selector string) ([]PropertyMetadata, error) {
	return selectorMetadataByFunc(
		optionTemplates,
		selector,
		func(optionTemplate OptionTemplate) string {
			return optionTemplate.SelectValue
		})
}

func selectorMetadataByFunc(optionTemplates []OptionTemplate, selector string, matchFunc func(optionTemplate OptionTemplate) string) ([]PropertyMetadata, error) {
	var options []string
	for _, optionTemplate := range optionTemplates {
		match := matchFunc(optionTemplate)
		options = append(options, match)

		if strings.EqualFold(selector, match) {
			return optionTemplate.PropertyMetadata, nil
		}
	}
	fmt.Println(fmt.Sprintf("Option template not found for selector [%s] options include %v", selector, options))
	return nil, nil
}
