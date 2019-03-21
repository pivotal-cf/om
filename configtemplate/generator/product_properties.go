package generator

import (
	"fmt"
	"strings"
)

func CreateProductProperties(metadata *Metadata) (map[string]PropertyValue, error) {
	productProperties := make(map[string]PropertyValue)
	for _, property := range metadata.Properties() {
		propertyMetadata, err := metadata.GetPropertyMetadata(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsConfigurable() && propertyMetadata.IsRequired() {
			if propertyMetadata.IsDropdown() {
				continue
			}
			if propertyMetadata.IsCollection() {
				if propertyMetadata.IsRequiredCollection() {
					propertyType, err := CollectionPropertyType(strings.Replace(property.Reference, ".", "", 1), propertyMetadata.Default, propertyMetadata.PropertyMetadata)
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
			for _, selector := range property.Selectors {
				if strings.EqualFold(defaultSelector, selector.Reference) {
					selectorMetadata := SelectorMetadataBySelectValue(propertyMetadata.OptionTemplates, fmt.Sprintf("%s", propertyMetadata.Default))
					for _, metadata := range selectorMetadata {
						if metadata.IsExplicityConfigurable() && metadata.IsRequired() && !metadata.IsDropdown() {
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

func CreateProductPropertiesVars(metadata *Metadata) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	for _, property := range metadata.Properties() {
		propertyMetadata, err := metadata.GetPropertyMetadata(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsMultiSelect() {
			continue
		}
		if propertyMetadata.IsConfigurable() && propertyMetadata.IsRequired() {
			if propertyMetadata.IsCollection() {
				if propertyMetadata.IsRequiredCollection() {
					CollectionPropertyVars(strings.Replace(property.Reference, ".", "", 1), propertyMetadata.PropertyMetadata, vars)
				}
			} else {
				if !propertyMetadata.IsSelector() {
					addPropertyToVars(property.Reference, propertyMetadata, vars)
				}
			}
		}

		if propertyMetadata.IsSelector() {
			defaultSelector := propertyMetadata.DefaultSelectorPath(property.Reference)
			for _, selector := range property.Selectors {
				if strings.EqualFold(defaultSelector, selector.Reference) {
					selectorMetadata := SelectorMetadata(propertyMetadata.OptionTemplates, fmt.Sprintf("%s", propertyMetadata.DefaultSelector()))
					for _, metadata := range selectorMetadata {
						if metadata.IsConfigurable() {
							selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, metadata.Name)
							addPropertyToVars(selectorProperty, &metadata, vars)
						}
					}
				}
			}
		}
	}
	return vars, nil
}

func addPropertyToVars(propertyName string, propertyMetadata *PropertyMetadata, vars map[string]interface{}) {
	if !propertyMetadata.IsSecret() {
		newPropertyName := strings.Replace(propertyName, ".", "", 1)
		newPropertyName = strings.Replace(newPropertyName, "properties.", "", 1)
		newPropertyName = strings.Replace(newPropertyName, ".", "/", -1)
		if propertyMetadata.Default != nil {
			if propertyMetadata.IsMultiSelect() {
				if _, ok := propertyMetadata.Default.([]interface{}); ok {
					vars[newPropertyName] = propertyMetadata.Default
				}
			} else {
				vars[newPropertyName] = propertyMetadata.Default
			}
		} else if propertyMetadata.IsBool() {
			vars[newPropertyName] = false
		}
	}
}

func CreateOpsFileName(propertyKey string) string {
	opsFileName := strings.Replace(propertyKey, "properties.", "", 1)
	opsFileName = strings.Replace(opsFileName, ".", "-", -1)
	return opsFileName
}

func CreateProductPropertiesOptionalOpsFiles(metadata *Metadata) (map[string][]Ops, error) {
	opsFiles := make(map[string][]Ops)
	for _, property := range metadata.Properties() {
		propertyKey := strings.Replace(property.Reference, ".", "", 1)
		opsFileName := CreateOpsFileName(propertyKey)
		propertyMetadata, err := metadata.GetPropertyMetadata(property.Reference)
		if err != nil {
			return nil, err
		}
		if propertyMetadata.IsConfigurable() {
			if propertyMetadata.IsSelector() {
				for _, selector := range property.Selectors {
					selectorMetadata := SelectorMetadata(propertyMetadata.OptionTemplates, strings.Replace(selector.Reference, property.Reference+".", "", 1))
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
								Value: CollectionOpsFile(i, propertyKey, propertyMetadata.PropertyMetadata),
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

func CreateProductPropertiesFeaturesOpsFiles(metadata *Metadata) (map[string][]Ops, error) {
	opsFiles := make(map[string][]Ops)
	for _, property := range metadata.Properties() {
		propertyMetadata, err := metadata.GetPropertyMetadata(property.Reference)
		if err != nil {
			return nil, err
		}

		if propertyMetadata.IsMultiSelect() && len(propertyMetadata.Options) > 1 {
			multiselectOpsFiles(property.Reference, propertyMetadata, opsFiles)
		}

		if propertyMetadata.IsSelector() {
			defaultSelector := propertyMetadata.DefaultSelectorPath(property.Reference)
			for _, selector := range property.Selectors {
				if !strings.EqualFold(defaultSelector, selector.Reference) {
					var ops []Ops
					opsFileName := strings.Replace(selector.Reference, ".", "", 1)
					opsFileName = strings.Replace(opsFileName, "properties.", "", 1)
					opsFileName = strings.Replace(opsFileName, ".", "-", -1)

					optionTemplate := propertyMetadata.OptionTemplate(strings.Replace(selector.Reference, property.Reference+".", "", 1))
					if optionTemplate != nil {
						ops = append(ops,
							Ops{
								Type: "replace",
								Path: fmt.Sprintf("/product-properties/%s?", property.Reference),
								Value: &OpsValue{
									Value:          optionTemplate.SelectValue,
									SelectedOption: optionTemplate.Name,
								},
							},
						)
					}

					if propertyMetadata.Default != nil {
						defaultSelectorMetadata := SelectorMetadataBySelectValue(propertyMetadata.OptionTemplates, fmt.Sprintf("%s", propertyMetadata.Default))
						for _, metadata := range defaultSelectorMetadata {
							selectorProperty := fmt.Sprintf("%s.%s", defaultSelector, metadata.Name)
							ops = append(ops,
								Ops{
									Type: "remove",
									Path: fmt.Sprintf("/product-properties/%s?", selectorProperty),
								},
							)
						}
					}
					selectorParts := strings.Split(selector.Reference, ".")
					selectorMetadata := SelectorMetadata(propertyMetadata.OptionTemplates, selectorParts[len(selectorParts)-1])
					for _, metadata := range selectorMetadata {
						if metadata.IsMultiSelect() && len(metadata.Options) > 1 {
							multiselectOpsFiles(selector.Reference+"."+metadata.Name, &metadata, opsFiles)
						}
						if metadata.IsConfigurable() && metadata.IsRequired() && !metadata.IsDropdown() {
							selectorProperty := fmt.Sprintf("%s.%s", selector.Reference, metadata.Name)
							ops = append(ops,
								Ops{
									Type:  "replace",
									Path:  fmt.Sprintf("/product-properties/%s?", selectorProperty),
									Value: metadata.PropertyType(strings.Replace(selectorProperty, ".", "", 1)),
								},
							)
						}
					}
					opsFiles[opsFileName] = ops
				}
			}
		}
	}
	return opsFiles, nil
}

func multiselectOpsFiles(propertyName string, propertyMetadata *PropertyMetadata, opsFiles map[string][]Ops) {
	for _, option := range propertyMetadata.Options {
		opsFileName := CreateOpsFileName(strings.Replace(propertyName, ".", "", 1))
		opsFileName = fmt.Sprintf("%s_%v", opsFileName, option.Name)

		opsFiles[opsFileName] = []Ops{
			{
				Type:  "replace",
				Path:  fmt.Sprintf("/product-properties/%s?/value/-", propertyName),
				Value: StringOpsValue(option.Name.(string)),
			},
		}
	}
}
