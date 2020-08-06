package api

import (
	"fmt"
	"regexp"
	"sort"
)

type updatedPropertyCollectionItem struct {
	Data map[string]interface{}
}
type updatedPropertyCollection []updatedPropertyCollectionItem

func (item updatedPropertyCollectionItem) getFieldValue(fieldName string) string {
	if value, ok := item.Data[fieldName].(string); ok {
		return value
	}
	return ""
}

func (item updatedPropertyCollectionItem) setFieldValue(fieldName string, value string) {
	item.Data[fieldName] = value
}

func parseUpdatedPropertyCollection(updatedProperty interface{}) (updatedPropertyCollection, error) {
	var collection updatedPropertyCollection

	updatedPropertyAsMap, ok := updatedProperty.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parseUpdatedPropertyCollection: failed to convert %v to map[string]interface{}", updatedProperty)
	}

	rawItems := updatedPropertyAsMap["value"]

	rawItemSlice, ok := rawItems.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parseUpdatedPropertyCollection: failed to convert %v to []interface{}", rawItems)
	}

	for _, item := range rawItemSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("parseUpdatedPropertyCollection: failed to convert %v to map[string]interface{}", item)
		}

		collection = append(collection, updatedPropertyCollectionItem{Data: itemMap})
	}

	return collection, nil
}

type responsePropertyCollection []responsePropertyCollectionItem
type responsePropertyCollectionItem struct {
	Data map[interface{}]interface{}
}

func parseResponsePropertyCollection(rawItems interface{}) (responsePropertyCollection, error) {
	var collection responsePropertyCollection

	rawItemSlice, ok := rawItems.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parseResponsePropertyCollection: failed to convert %v to []interface{}", rawItems)
	}

	for _, item := range rawItemSlice {
		itemMap, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("parseResponsePropertyCollection: failed to convert %v to map[interface{}]interface{}", item)
		}

		collection = append(collection, responsePropertyCollectionItem{Data: itemMap})
	}

	return collection, nil
}

func (item responsePropertyCollectionItem) getFieldValue(fieldName string) string {
	if valueObj, ok := item.Data[fieldName].(map[interface{}]interface{}); ok {
		if value, ok := valueObj["value"].(string); ok {
			return value
		}
	}
	return ""
}

func (item responsePropertyCollectionItem) getFieldValuesExceptGUID() map[interface{}]interface{} {
	extractedValues := make(map[interface{}]interface{})

	for key, valueObj := range item.Data {
		if key == "guid" {
			continue
		}
		extractedValues[key] = valueObj.(map[interface{}]interface{})["value"]
	}

	return extractedValues
}

func (item responsePropertyCollectionItem) getSortedFieldNames() []string {
	sortedFieldNames := make([]string, 0, len(item.Data))
	for k := range item.Data {
		sortedFieldNames = append(sortedFieldNames, fmt.Sprintf("%v", k))
	}
	sort.Strings(sortedFieldNames)

	return sortedFieldNames
}

func (item responsePropertyCollectionItem) findLogicalKeyField() (string, bool) {
	//search order is important; 'name' should be found before 'ending-with-name'
	regexes := []string{"^name$", "^key$", "(?i)name$"}
	for _, regex := range regexes {
		compiledRegex := regexp.MustCompile(regex)
		for _, fieldName := range item.getSortedFieldNames() {
			if compiledRegex.MatchString(fieldName) {
				return fieldName, true
			}
		}
	}

	return "", false
}

func assignExistingGUIDUsingLogicalKey(updatedCollection updatedPropertyCollection, existingCollection responsePropertyCollection) bool {
	//since all collection items have the same structure; only need to find the logical key from the first one
	logicalKeyFieldName, found := existingCollection[0].findLogicalKeyField()
	if !found {
		return false
	}

	for _, updatedCollectionItem := range updatedCollection {
		updatedCollectionItemLogicalKeyValue := updatedCollectionItem.getFieldValue(logicalKeyFieldName)
		for _, existingCollectionItem := range existingCollection {
			if updatedCollectionItemLogicalKeyValue == existingCollectionItem.getFieldValue(logicalKeyFieldName) {
				updatedCollectionItem.setFieldValue("guid", existingCollectionItem.getFieldValue("guid"))
				break //move onto the next updatedCollectionItem
			}
		}
	}

	return true
}

func (existingCollection responsePropertyCollection) findGUIDForIEquivalentlItem(updatedProperty updatedPropertyCollectionItem) (string, bool) {

	for _, existingCollectionItem := range existingCollection {
		//use the fact that fmt prints maps in order (see: https://tip.golang.org/doc/go1.12#fmt) to check for equivalence
		if fmt.Sprintf("%+v", updatedProperty.Data) == fmt.Sprintf("%+v", existingCollectionItem.getFieldValuesExceptGUID()) {
			return existingCollectionItem.getFieldValue("guid"), true
		}
	}

	return "", false
}

func assignExistingGUIDUsingEquivalentValue(updatedCollection updatedPropertyCollection, existingCollection responsePropertyCollection) bool {

	foundEquivalentItems := false

	for _, updatedCollectionItem := range updatedCollection {
		if guid, ok := existingCollection.findGUIDForIEquivalentlItem(updatedCollectionItem); ok {
			updatedCollectionItem.setFieldValue("guid", guid)
			foundEquivalentItems = true
		}
	}

	return foundEquivalentItems
}

func associateExistingCollectionGUIDs(updatedProperty interface{}, existingProperty ResponseProperty) error {
	updatedCollection, err := parseUpdatedPropertyCollection(updatedProperty)
	if err != nil {
		return err
	}
	existingCollection, err := parseResponsePropertyCollection(existingProperty.Value)
	if err != nil {
		return err
	}

	if assignExistingGUIDUsingEquivalentValue(updatedCollection, existingCollection) {
		return nil
	}

	if assignExistingGUIDUsingLogicalKey(updatedCollection, existingCollection) {
		return nil
	}

	return nil
}
