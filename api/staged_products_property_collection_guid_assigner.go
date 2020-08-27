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

func (item updatedPropertyCollectionItem) getSortedFieldNames() []string {
	sortedFieldNames := make([]string, 0, len(item.Data))
	for k := range item.Data {
		sortedFieldNames = append(sortedFieldNames, fmt.Sprintf("%v", k))
	}
	sort.Strings(sortedFieldNames)

	return sortedFieldNames
}

func (item updatedPropertyCollectionItem) findLogicalKeyField() (string, bool) {
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

func parseResponsePropertyCollection(input associateExistingCollectionGUIDsInput) (responsePropertyCollection, error) {
	var collection responsePropertyCollection

	rawItemSlice, ok := input.ExistingProperty.Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("parseResponsePropertyCollection: failed to convert %v to []interface{}", input.ExistingProperty.Value)
	}

	for index, item := range rawItemSlice {
		itemMap, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("parseResponsePropertyCollection: failed to convert %v to map[interface{}]interface{}", item)
		}

		for collectionItemKey, collectionItemObj := range itemMap {
			collectionItemObjAsMap := collectionItemObj.(map[interface{}]interface{})
			isCredential := collectionItemObjAsMap["credential"].(bool)
			if !isCredential {
				continue
			}

			credentialName := fmt.Sprintf("%s[%d].%s", input.PropertyName, index, collectionItemKey)
			apiOutput, err := input.APIService.GetDeployedProductCredential(GetDeployedProductCredentialInput{
				DeployedGUID:        input.ProductGUID,
				CredentialReference: credentialName,
			})
			if err != nil {
				return nil, err
			}

			collectionItemObjAsMap["value"] = apiOutput.Credential.Value
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
		isNotConfigurable := !valueObj.(map[interface{}]interface{})["configurable"].(bool)
		if isNotConfigurable {
			continue
		}
		extractedValues[key] = valueObj.(map[interface{}]interface{})["value"]
	}

	return extractedValues
}

func (existingCollection responsePropertyCollection) findGUIDForEquivalentlItem(updatedProperty updatedPropertyCollectionItem) (string, bool) {

	for _, existingCollectionItem := range existingCollection {
		//use the fact that fmt prints maps in order (see: https://tip.golang.org/doc/go1.12#fmt) to check for equivalence
		if fmt.Sprintf("%+v", updatedProperty.Data) == fmt.Sprintf("%+v", existingCollectionItem.getFieldValuesExceptGUID()) {
			return existingCollectionItem.getFieldValue("guid"), true
		}
	}

	return "", false
}

type associateExistingCollectionGUIDsInput struct {
	APIService       Api
	ProductGUID      string
	PropertyName     string
	UpdatedProperty  interface{}
	ExistingProperty ResponseProperty
}

func associateExistingCollectionGUIDs(input associateExistingCollectionGUIDsInput) error {
	updatedCollection, err := parseUpdatedPropertyCollection(input.UpdatedProperty)
	if err != nil {
		return err
	}
	existingCollection, err := parseResponsePropertyCollection(input)
	if err != nil {
		return err
	}

	for _, updatedCollectionItem := range updatedCollection {
		if guid, ok := existingCollection.findGUIDForEquivalentlItem(updatedCollectionItem); ok {
			updatedCollectionItem.setFieldValue("guid", guid)
		} else if logicalKeyFieldName, ok := updatedCollectionItem.findLogicalKeyField(); ok {
			updatedCollectionItemLogicalKeyValue := updatedCollectionItem.getFieldValue(logicalKeyFieldName)
			for _, existingCollectionItem := range existingCollection {
				if updatedCollectionItemLogicalKeyValue == existingCollectionItem.getFieldValue(logicalKeyFieldName) {
					updatedCollectionItem.setFieldValue("guid", existingCollectionItem.getFieldValue("guid"))
				}
			}
		}

	}

	return nil
}
