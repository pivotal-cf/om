package api

import "fmt"

var errNotFound = fmt.Errorf("element not found")

func getString(element interface{}, key string) (string, error) {
	value, err := get(element, key)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", errNotFound
	}
	strVal, ok := value.(string)
	if ok {
		return strVal, nil
	}
	return "", fmt.Errorf("element %v (%T) with key %q is not a string", element, value, key)
}

func set(element interface{}, key, value string) error {
	mapString, ok := element.(map[string]interface{})
	if ok {
		mapString[key] = value
		return nil
	}
	mapInterface, ok := element.(map[interface{}]interface{})
	if ok {
		mapInterface[key] = value
		return nil
	}
	return fmt.Errorf("Unexpected type %v", element)
}

func get(element interface{}, key string) (interface{}, error) {
	mapString, ok := element.(map[string]interface{})
	if ok {
		return mapString[key], nil
	}
	mapInterface, ok := element.(map[interface{}]interface{})
	if ok {
		return mapInterface[key], nil
	}
	return nil, fmt.Errorf("Unexpected type %v", element)
}

func collectionElementGUID(propertyName, elementName string, configuredProperties map[string]ResponseProperty) (string, error) {
	collection := configuredProperties[propertyName].Value
	collectionArray := collection.([]interface{})
	for _, collectionElement := range collectionArray {
		element, err := get(collectionElement, "name")
		if err != nil {
			return "", err
		}
		currentElement, err := getString(element, "value")
		if err != nil {
			return "", err
		}
		if currentElement == elementName {
			guidElement, err := get(collectionElement, "guid")
			if err != nil {
				return "", err
			}
			guid, err := getString(guidElement, "value")
			if err != nil {
				return "", err
			}
			return guid, nil
		}

	}
	return "", nil
}

//Find and associate the GUID for those collection items that already exist in OpsMgr
//This ensures that updates to existing collection items don't trigger deletion & recreation (with a new GUID)
func associateExistingCollectionGUIDs(property interface{}, propertyName string, currentConfiguredProperties map[string]ResponseProperty) error {
	collectionValue, err := get(property, "value")
	if err != nil {
		return err
	}
	for _, collectionElement := range collectionValue.([]interface{}) {
		name, err := getString(collectionElement, "name")
		if err != nil {
			if err == errNotFound {
				continue
			}
			return err
		}
		guid, err := collectionElementGUID(propertyName, name, currentConfiguredProperties)
		if err != nil {
			return err
		}
		err = set(collectionElement, "guid", guid)
		if err != nil {
			return err
		}
	}
	return nil
}
