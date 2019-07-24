package generator

import (
	"fmt"
	"strings"
)

func GetAllProductProperties(metadata *Metadata) (map[string]PropertyValue, error) {
	productProperties := make(map[string]PropertyValue)
	for _, property := range metadata.PropertyInputs() {
		propertyMetadata, err := metadata.GetPropertyBlueprint(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsConfigurable() && propertyMetadata.IsRequired() {
			if propertyMetadata.IsDropdown() {
				continue
			}
			if propertyMetadata.IsCollection() {
				if propertyMetadata.IsRequiredCollection() {
					propertyType, err := collectionPropertyType(strings.Replace(property.Reference, ".", "", 1), propertyMetadata.Default, propertyMetadata.PropertyBlueprints)
					if err != nil {
						return nil, err
					}
					if propertyType != nil {
						productProperties[property.Reference] = propertyType
					}
				}
			} else {
				propertyType := propertyMetadata.PropertyType(strings.Replace(property.Reference, ".", "", 1))
				if propertyType != nil {
					productProperties[property.Reference] = propertyType
				}
			}
		}

		if propertyMetadata.IsSelector() {
			defaultSelector := propertyMetadata.DefaultSelectorPath(property.Reference)
			for _, selector := range property.SelectorPropertyInputs {
				if strings.EqualFold(defaultSelector, selector.Reference) {
					selectorMetadata := SelectorBlueprintsBySelectValue(propertyMetadata.OptionTemplates, fmt.Sprintf("%s", propertyMetadata.Default))
					for _, metadata := range selectorMetadata {
						if metadata.IsConfigurable() && metadata.IsRequired() && !metadata.IsDropdown() {
							selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, metadata.Name)
							propertyType := metadata.PropertyType(strings.Replace(selectorProperty, ".", "", 1))
							if propertyType != nil {
								productProperties[selectorProperty] = propertyType
							}
						}
					}
				}
			}
		}
	}
	return productProperties, nil
}

func GetDefaultPropertyVars(metadata *Metadata) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, property := range metadata.PropertyInputs() {
		propertyMetadata, err := metadata.GetPropertyBlueprint(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsMultiSelect() {
			continue
		}
		if propertyMetadata.IsConfigurable() && propertyMetadata.IsRequired() {
			if propertyMetadata.IsCollection() {
				collectionPropertyVars(strings.Replace(property.Reference, ".", "", 1), propertyMetadata.PropertyBlueprints, true, vars)
			} else {
				if !propertyMetadata.IsSelector() {
					addPropertyToVars(property.Reference, propertyMetadata, true, vars)
				}
			}
		}

		if propertyMetadata.IsSelector() {
			defaultSelector := propertyMetadata.DefaultSelectorPath(property.Reference)
			for _, selector := range property.SelectorPropertyInputs {
				if strings.EqualFold(defaultSelector, selector.Reference) {
					selectorMetadata := SelectorOptionsBlueprints(propertyMetadata.OptionTemplates, fmt.Sprintf("%s", propertyMetadata.DefaultSelector()))
					for _, metadata := range selectorMetadata {
						if metadata.IsConfigurable() {
							selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, metadata.Name)
							addPropertyToVars(selectorProperty, &metadata, true, vars)
						}
					}
				}
			}
		}
	}
	return vars, nil
}

func GetRequiredPropertyVars(metadata *Metadata) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, property := range metadata.PropertyInputs() {
		propertyBlueprint, err := metadata.GetPropertyBlueprint(property.Reference)
		if err != nil {
			return nil, fmt.Errorf("could not create required-vars file: %s", err.Error())
		}

		if propertyBlueprint.IsMultiSelect() || !propertyBlueprint.IsConfigurable() || !propertyBlueprint.IsRequired() || propertyBlueprint.IsDropdown() {
			continue
		}

		if propertyBlueprint.IsCollection() {
			collectionPropertyVars(strings.Replace(property.Reference, ".", "", 1), propertyBlueprint.PropertyBlueprints, false, vars)
			continue
		}

		if propertyBlueprint.IsSelector() && propertyBlueprint.HasDefault() {
			defaultSelector := propertyBlueprint.DefaultSelectorPath(property.Reference)
			for _, selector := range property.SelectorPropertyInputs {
				if !strings.EqualFold(defaultSelector, selector.Reference) {
					continue
				}

				selectorOptionBlueprints := SelectorOptionsBlueprints(propertyBlueprint.OptionTemplates, fmt.Sprintf("%s", propertyBlueprint.DefaultSelector()))
				for _, selectorOptionBlueprint := range selectorOptionBlueprints {
					if selectorOptionBlueprint.IsConfigurable() && selectorOptionBlueprint.IsRequired() && !selectorOptionBlueprint.IsMultiSelect() {
						selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, selectorOptionBlueprint.Name)
						addPropertyToVars(selectorProperty, &selectorOptionBlueprint, false, vars)
					}
				}
			}
			continue
		}

		addPropertyToVars(property.Reference, propertyBlueprint, false, vars)
	}
	return vars, nil
}

func addPropertyToVars(propertyName string, propertyBlueprint *PropertyBlueprint, includePropertiesWithDefaults bool, vars map[string]interface{}) {
	newPropertyName := strings.Replace(propertyName, ".", "", 1)
	newPropertyName = strings.Replace(newPropertyName, "properties.", "", 1)
	newPropertyName = strings.Replace(newPropertyName, ".", "_", -1)
	if !propertyBlueprint.IsSecret() && !propertyBlueprint.IsSimpleCredentials() && !propertyBlueprint.IsCertificate() {
		if includePropertiesWithDefaults {
			if propertyBlueprint.HasDefault() {
				if propertyBlueprint.IsMultiSelect() {
					if _, ok := propertyBlueprint.Default.([]interface{}); ok {
						vars[newPropertyName] = propertyBlueprint.Default
						return
					}
				}

				vars[newPropertyName] = propertyBlueprint.Default
				return
			} else if propertyBlueprint.IsBool() {
				vars[newPropertyName] = false
				return
			}

			return
		}

		if !propertyBlueprint.HasDefault() && !propertyBlueprint.IsBool() {
			vars[newPropertyName] = ""
		}

		return
	}

	if !includePropertiesWithDefaults && !propertyBlueprint.HasDefault() {
		if propertyBlueprint.IsCertificate() {
			vars[fmt.Sprintf("%s_%s", newPropertyName, "certificate")] = ""
			vars[fmt.Sprintf("%s_%s", newPropertyName, "privatekey")] = ""

			return
		}

		vars[newPropertyName] = ""
	}
}

func CreateOpsFileName(propertyKey string) string {
	opsFileName := strings.Replace(propertyKey, "properties.", "", 1)
	opsFileName = strings.Replace(opsFileName, ".", "-", -1)
	return opsFileName
}

func CreateProductPropertiesOptionalOpsFiles(metadata *Metadata) (map[string][]Ops, error) {
	opsFiles := make(map[string][]Ops)
	for _, property := range metadata.PropertyInputs() {
		propertyKey := strings.Replace(property.Reference, ".", "", 1)
		opsFileName := CreateOpsFileName(propertyKey)
		propertyMetadata, err := metadata.GetPropertyBlueprint(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsConfigurable() {
			if propertyMetadata.IsSelector() {
				for _, selector := range property.SelectorPropertyInputs {
					selectorMetadata := SelectorOptionsBlueprints(propertyMetadata.OptionTemplates, strings.Replace(selector.Reference, property.Reference+".", "", 1))
					for _, metadata := range selectorMetadata {
						if metadata.IsConfigurable() {
							if !metadata.IsRequired() || metadata.IsDropdown() {
								selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, metadata.Name)
								opsFiles[fmt.Sprintf("add-%s", CreateOpsFileName(strings.Replace(selectorProperty, ".", "", 1)))] = []Ops{
									{
										Type:  "replace",
										Path:  fmt.Sprintf("/product-properties/%s?", selectorProperty),
										Value: metadata.PropertyType(strings.Replace(selectorProperty, ".", "", 1)),
									},
								}
							}
						}
					}
				}
			} else {
				if propertyMetadata.IsDropdown() {
					opsFiles[fmt.Sprintf("add-%s", opsFileName)] = []Ops{
						{
							Type:  "replace",
							Path:  fmt.Sprintf("/product-properties/%s?", property.Reference),
							Value: propertyMetadata.PropertyType(propertyKey),
						},
					}
				} else if propertyMetadata.IsCollection() {
					x := 1
					if propertyMetadata.IsRequired() {
						x = 2
					}
					for i := x; i <= 10; i++ {
						opsFiles[fmt.Sprintf("add-%d-%s", i, opsFileName)] = []Ops{
							{
								Type:  "replace",
								Path:  fmt.Sprintf("/product-properties/%s?", property.Reference),
								Value: collectionOpsFile(i, propertyKey, propertyMetadata.PropertyBlueprints),
							},
						}
					}
				} else if !propertyMetadata.IsRequired() {
					opsFiles[fmt.Sprintf("add-%s", opsFileName)] = []Ops{
						{
							Type:  "replace",
							Path:  fmt.Sprintf("/product-properties/%s?", property.Reference),
							Value: propertyMetadata.PropertyType(strings.Replace(property.Reference, ".", "", 1)),
						},
					}
				}
			}
		}
	}

	return opsFiles, nil
}

func CreateProductPropertiesFeaturesOpsFiles(tileMetadata *Metadata) (map[string][]Ops, error) {
	opsFiles := make(map[string][]Ops)
	for _, propertyReference := range tileMetadata.PropertyInputs() {
		propertyBlueprint, err := tileMetadata.GetPropertyBlueprint(propertyReference.Reference)
		if err != nil {
			return nil, fmt.Errorf("could not create feature ops files: %s", err.Error())
		}

		if propertyBlueprint.IsMultiSelect() && len(propertyBlueprint.Options) > 1 {
			multiselectOpsFiles(propertyReference.Reference, propertyBlueprint, opsFiles)
		}

		if propertyBlueprint.IsSelector() {
			defaultSelector := propertyBlueprint.DefaultSelectorPath(propertyReference.Reference)
			for _, selectorPropertyInput := range propertyReference.SelectorPropertyInputs {
				if !strings.EqualFold(defaultSelector, selectorPropertyInput.Reference) {
					var ops []Ops
					featureOpsFileName := strings.Replace(selectorPropertyInput.Reference, ".", "", 1)
					featureOpsFileName = strings.Replace(featureOpsFileName, "properties.", "", 1)
					featureOpsFileName = strings.Replace(featureOpsFileName, ".", "-", -1)

					optionTemplate := propertyBlueprint.OptionTemplate(strings.Replace(selectorPropertyInput.Reference, propertyReference.Reference+".", "", 1))
					if optionTemplate != nil {
						ops = append(ops,
							Ops{
								Type: "replace",
								Path: fmt.Sprintf("/product-properties/%s?", propertyReference.Reference),
								Value: &OpsValue{
									Value:          optionTemplate.SelectValue,
									SelectedOption: optionTemplate.Name,
								},
							},
						)
					}

					if propertyBlueprint.HasDefault() {
						defaultSelectorBlueprints := SelectorBlueprintsBySelectValue(propertyBlueprint.OptionTemplates, fmt.Sprintf("%s", propertyBlueprint.Default))
						for _, selectorBlueprint := range defaultSelectorBlueprints {
							selectorProperty := fmt.Sprintf("%s.%s", defaultSelector, selectorBlueprint.Name)
							ops = append(ops,
								Ops{
									Type: "remove",
									Path: fmt.Sprintf("/product-properties/%s?", selectorProperty),
								},
							)
						}
					}
					selectorReferenceParts := strings.Split(selectorPropertyInput.Reference, ".")
					selectorBlueprints := SelectorOptionsBlueprints(propertyBlueprint.OptionTemplates, selectorReferenceParts[len(selectorReferenceParts)-1])
					for _, selectorBlueprint := range selectorBlueprints {
						if selectorBlueprint.IsMultiSelect() && len(selectorBlueprint.Options) > 1 {
							multiselectOpsFiles(selectorPropertyInput.Reference+"."+selectorBlueprint.Name, &selectorBlueprint, opsFiles)
						}
						if selectorBlueprint.IsConfigurable() && selectorBlueprint.IsRequired() && !selectorBlueprint.IsDropdown() {
							selectorProperty := fmt.Sprintf("%s.%s", selectorPropertyInput.Reference, selectorBlueprint.Name)
							ops = append(ops,
								Ops{
									Type:  "replace",
									Path:  fmt.Sprintf("/product-properties/%s?", selectorProperty),
									Value: selectorBlueprint.PropertyType(strings.Replace(selectorProperty, ".", "", 1)),
								},
							)
						}
					}
					opsFiles[featureOpsFileName] = ops
				}
			}
		}
	}

	return opsFiles, nil
}

func multiselectOpsFiles(propertyName string, propertyBlueprint *PropertyBlueprint, opsFiles map[string][]Ops) {
	for _, option := range propertyBlueprint.Options {
		multiSelectOpsFileName := CreateOpsFileName(strings.Replace(propertyName, ".", "", 1))
		multiSelectOpsFileName = fmt.Sprintf("%s_%v", multiSelectOpsFileName, option.Name)

		opsFiles[multiSelectOpsFileName] = []Ops{
			{
				Type:  "replace",
				Path:  fmt.Sprintf("/product-properties/%s?/value/-", propertyName),
				Value: StringOpsValue(option.Name.(string)),
			},
		}
	}
}
