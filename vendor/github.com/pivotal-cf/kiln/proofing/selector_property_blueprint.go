package proofing

import "fmt"

type SelectorPropertyBlueprint struct {
	SimplePropertyBlueprint `yaml:",inline"`

	OptionTemplates []SelectorPropertyOptionTemplate `yaml:"option_templates"`

	// TODO: validations: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/selector_property_blueprint.rb#L10
	// TODO: find_object: https://github.com/pivotal-cf/installation/blob/039a2ef3f751ef5915c425da8150a29af4b764dd/web/app/models/persistence/metadata/selector_property_blueprint.rb#L8
}

func (sp SelectorPropertyBlueprint) Normalize(prefix string) []NormalizedPropertyBlueprint {
	propertyBlueprints := sp.SimplePropertyBlueprint.Normalize(prefix)

	for _, optionTemplate := range sp.OptionTemplates {
		for _, otpb := range optionTemplate.PropertyBlueprints {
			propertyName := fmt.Sprintf("%s.%s.%s", prefix, sp.Name, optionTemplate.Name)
			propertyBlueprints = append(propertyBlueprints, otpb.Normalize(propertyName)...)
		}
	}

	return propertyBlueprints
}
