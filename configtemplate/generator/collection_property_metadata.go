package generator

import (
	"fmt"
	"reflect"
	"strings"
)

func IsDefaultAnArray(defaultValue interface{}) bool {
	if defaultValue == nil {
		return false
	}
	_, ok := defaultValue.([]interface{})
	return ok
}

func DefaultsArrayToCollectionArray(propertyName string, defaultValue interface{}, subProperties []PropertyMetadata) ([]map[string]SimpleType, error) {
	var collectionProperties []map[string]SimpleType
	for _, defaultValues := range defaultValue.([]interface{}) {
		arrayProperties := make(map[string]SimpleType)
		defaultMap := defaultValues.(map[interface{}]interface{})
		for key, value := range defaultMap {
			keyAsString := key.(string)
			if value != nil {
				switch value.(type) {
				case string:
					arrayProperties[keyAsString] = SimpleString(value.(string))
				case bool:
					arrayProperties[keyAsString] = SimpleBoolean(value.(bool))
				case int:
					arrayProperties[keyAsString] = SimpleInteger(value.(int))

				default:
					return nil, fmt.Errorf("value %v is not known", reflect.TypeOf(value))
				}
			}
		}
		for _, subProperty := range subProperties {
			if _, ok := arrayProperties[subProperty.Name]; !ok {
				arrayProperties[subProperty.Name] = SimpleString(fmt.Sprintf("((%s/%s))", propertyName, subProperty.Name))
			}
		}
		collectionProperties = append(collectionProperties, arrayProperties)
	}
	return collectionProperties, nil
}

func DefaultsToArray(propertyName string, subProperties []PropertyMetadata) map[string]SimpleType {
	properties := make(map[string]SimpleType)
	for _, subProperty := range subProperties {
		if subProperty.IsConfigurable() {
			if subProperty.IsSecret() {
				properties[subProperty.Name] = &SecretValue{
					Value: fmt.Sprintf("((%s/%s))", propertyName, subProperty.Name),
				}
			} else if subProperty.IsCertificate() {
				properties[subProperty.Name] = NewCertificateValue(propertyName)
			} else {
				properties[subProperty.Name] = SimpleString(fmt.Sprintf("((%s/%s))", propertyName, subProperty.Name))
			}
		}
	}
	return properties
}

func CollectionPropertyType(propertyName string, defaultValue interface{}, subProperties []PropertyMetadata) (PropertyValue, error) {
	propertyName = strings.Replace(propertyName, "properties.", "", 1)
	propertyName = fmt.Sprintf("%s_0", strings.Replace(propertyName, ".", "/", -1))
	var collectionProperties []map[string]SimpleType
	if IsDefaultAnArray(defaultValue) {
		defaultArrayProperties, err := DefaultsArrayToCollectionArray(propertyName, defaultValue, subProperties)
		if err != nil {
			return nil, err
		}
		collectionProperties = append(collectionProperties, defaultArrayProperties...)
	} else {
		collectionProperties = append(collectionProperties, DefaultsToArray(propertyName, subProperties))
	}

	return &CollectionsPropertiesValueHolder{
		Value: collectionProperties,
	}, nil
}

func CollectionPropertyVars(propertyName string, subProperties []PropertyMetadata, vars map[string]interface{}) {
	propertyName = strings.Replace(propertyName, "properties.", "", 1)
	propertyName = fmt.Sprintf("%s_0", strings.Replace(propertyName, ".", "/", -1))
	for _, subProperty := range subProperties {
		if subProperty.IsConfigurable() {
			if !subProperty.IsSecret() && !subProperty.IsSimpleCredentials() && !subProperty.IsCertificate() {
				subPropertyName := fmt.Sprintf("%s/%s", propertyName, subProperty.Name)
				if subProperty.Default != nil {
					vars[subPropertyName] = subProperty.Default
				}
			}
		}
	}
}

func CollectionOpsFile(numOfElements int, propertyName string, subProperties []PropertyMetadata) OpsValueType {
	var collectionProperties []map[string]SimpleType
	for i := 1; i <= numOfElements; i++ {
		newPropertyName := strings.Replace(propertyName, "properties.", "", 1)
		newPropertyName = fmt.Sprintf("%s_%d", strings.Replace(newPropertyName, ".", "/", -1), i-1)
		properties := make(map[string]SimpleType)
		for _, subProperty := range subProperties {
			if subProperty.IsSecret() {
				properties[subProperty.Name] = &SecretValue{
					Value: fmt.Sprintf("((%s/%s))", newPropertyName, subProperty.Name),
				}
			} else if subProperty.IsCertificate() {
				properties[subProperty.Name] = NewCertificateValue(newPropertyName)
			} else {
				properties[subProperty.Name] = SimpleString(fmt.Sprintf("((%s/%s))", newPropertyName, subProperty.Name))
			}
		}
		collectionProperties = append(collectionProperties, properties)
	}
	return &CollectionsPropertiesValueHolder{
		Value: collectionProperties,
	}
}
