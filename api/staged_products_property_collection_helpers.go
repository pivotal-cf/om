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

//Find and return first string matching regex pattern
//returns matched <string, true> if found, or <"",false> if not
func findFirst(strings []string, fieldRegex string) (string, bool) {
	r, err := regexp.Compile(fieldRegex)
	if err != nil {
		panic(fmt.Sprintf("failed to compile regex: %v because:\n%v", fieldRegex, err))
	}
	for _, str := range strings {
		if r.MatchString(str) {
			return str, true
		}
	}
	return "", false
}

//Finds logical key field (if exists)
//returns <fieldName, true> if map contains field that can be considered a logical key
//returns <"", false> if no logical key found
func (item responsePropertyCollectionItem) findLogicalKeyField() (string, bool) {
	//Extract & sort the fieldnames to ensure search is deterministic (since map order isn't guarenteed)
	sortedFields := make([]string, 0, len(item.Data))
	for k := range item.Data {
		sortedFields = append(sortedFields, fmt.Sprintf("%v", k))
	}
	sort.Strings(sortedFields)

	if fieldName, ok := findFirst(sortedFields, "^name$"); ok { //First look for a field named 'name'
		return fieldName, true
	} else if fieldName, ok := findFirst(sortedFields, "^key$"); ok { //then a field named 'key'
		return fieldName, true
	} else if fieldName, ok := findFirst(sortedFields, "(?i)name$"); ok { // otherwise a field that ends with 'name' (case insensitive)
		return fieldName, true
	}

	return "", false
}

func (collection responsePropertyCollection) findLogicalKeyFieldFromFirstCollectionItem() (string, bool) {
	return collection[0].findLogicalKeyField()
}

//Find and associate the GUID for those collection items that already exist in OpsMgr
//This ensures that updates to existing collection items don't trigger deletion & recreation (with a new GUID)
func associateExistingCollectionGUIDs(updatedProperty interface{}, existingProperty ResponseProperty) error {
	updatedCollection, err := parseUpdatedPropertyCollection(updatedProperty)
	if err != nil {
		return err
	}
	existingCollection, err := parseResponsePropertyCollection(existingProperty.Value)
	if err != nil {
		return err
	}
	if logicalKeyFieldName, ok := existingCollection.findLogicalKeyFieldFromFirstCollectionItem(); ok {
		//Use the logical key to find associated GUIDs (from the existingProperty collection items) and assign them to the updatedProperty item
		for _, updatedCollectionItem := range updatedCollection {
			updatedCollectionItemLogicalKeyValue := updatedCollectionItem.getFieldValue(logicalKeyFieldName)
			for _, existingCollectionItem := range existingCollection {
				if updatedCollectionItemLogicalKeyValue == existingCollectionItem.getFieldValue(logicalKeyFieldName) {
					updatedCollectionItem.setFieldValue("guid", existingCollectionItem.getFieldValue("guid"))
					break //move onto the next updatedCollectionItem
				}
			}
		}
	}
	return nil
}
