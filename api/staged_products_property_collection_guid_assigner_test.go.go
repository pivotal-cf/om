package api //not in api_tests because we are intentionally testing some functionality internal to the package

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("ResponsePropertyCollection", func() {
	unmarshalJSONLikeApiGetStagedProductProperties := func(json string) interface{} {
		var rawCollection interface{}
		err := yaml.Unmarshal([]byte(json), &rawCollection)
		if err != nil {
			panic(fmt.Errorf("Failed to parse json: %w", err))
		}
		return rawCollection
	}
	unmarshalJSON := func(rawJSON string) interface{} {
		var rawCollection interface{}
		err := json.Unmarshal([]byte(rawJSON), &rawCollection)
		if err != nil {
			panic(fmt.Errorf("Failed to parse json: %w", err))
		}
		return rawCollection
	}
	When("parseResponsePropertyCollection", func() {
		It("parses all the elements in the collection", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-one",
						"optional": false
					},
					"some_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": "true",
						"optional": false
					}
				},
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"name": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_name",
						"optional": false
					},
					"yet_another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())
			Expect(len(collection)).To(Equal(2))
		})
	})
	When("extracting field values", func() {
		It("correctly extracts guids", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"name": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_name",
						"optional": false
					},
					"yet_another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())
			Expect(collection[0].getFieldValue("guid")).To(Equal("28bab1d3-4a4b-48d5-8dac-two"))
		})
		It("correctly extracts strings", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"name": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_name",
						"optional": false
					},
					"yet_another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())
			Expect(collection[0].getFieldValue("name")).To(Equal("the_name"))
		})
	})
	When("finding the logical key field", func() {
		It("finds a 'name' logical key", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"name": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_name",
						"optional": false
					},
					"yet_another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())

			key, ok := collection[0].findLogicalKeyField()
			Expect(ok).To(BeTrue())
			Expect(key).To(Equal("name"))
		})
		It("fails to find a logical key when there isn't one", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-one",
						"optional": false
					},
					"some_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": "true",
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())

			key, ok := collection[0].findLogicalKeyField()
			Expect(ok).To(BeFalse())
			Expect(key).To(BeEmpty())
		})

		It("finds a 'key' logical key", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-three",
						"optional": false
					},
					"key": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_key",
						"optional": false
					},
					"some_additional_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())

			key, ok := collection[0].findLogicalKeyField()
			Expect(ok).To(BeTrue())
			Expect(key).To(Equal("key"))
		})
		It("finds a logical key ending in 'name'", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-four",
						"optional": false
					},
					"sqlServerName": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_sqlserver_name",
						"optional": false
					},
					"some_additional_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": false,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())

			key, ok := collection[0].findLogicalKeyField()
			Expect(ok).To(BeTrue())
			Expect(key).To(Equal("sqlServerName"))
		})
		It("picks 'name' as the logical key when there is a 'name' field AND a field that ends in 'name' (eg: Filename)", func() {
			collection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-five",
						"optional": false
					},
					"name": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "the_name",
						"optional": false
					},
					"Filename": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "important_data.tgz",
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())

			key, ok := collection[0].findLogicalKeyField()
			Expect(ok).To(BeTrue())
			Expect(key).To(Equal("name"))
		})
	})
	When("matching based on item contents", func() {
		It("finds items that are identical", func() {
			existingCollection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-one",
						"optional": false
					},
					"a_property": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "\"value\" of 'first' item",
						"optional": false
					},
					"another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": true,
						"optional": false
					}
				},
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"a_property": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "\"value\" of 'second' item",
						"optional": false
					},
					"another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": true,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())
			updatedCollection, err := parseUpdatedPropertyCollection(unmarshalJSON(`{ "value":[
				{
					"a_property": "\"value\" of 'second' item",
					"another_property": true
				}
			]}`))
			Expect(err).To(BeNil())

			guid, ok := existingCollection.findGUIDForIEquivalentlItem(updatedCollection[0])
			Expect(ok).To(BeTrue())
			Expect(guid).To(Equal("28bab1d3-4a4b-48d5-8dac-two"))
		})
		It("finds items that are equivalent but have a different key order", func() {
			existingCollection, err := parseResponsePropertyCollection(unmarshalJSONLikeApiGetStagedProductProperties(`[
				{
					"guid": {
						"type": "uuid",
						"configurable": false,
						"credential": false,
						"value": "28bab1d3-4a4b-48d5-8dac-two",
						"optional": false
					},
					"a_property": {
						"type": "string",
						"configurable": true,
						"credential": false,
						"value": "\"value\" of 'second' item",
						"optional": false
					},
					"another_property": {
						"type": "boolean",
						"configurable": true,
						"credential": false,
						"value": true,
						"optional": false
					}
				}
			]`))
			Expect(err).To(BeNil())
			updatedCollection, err := parseUpdatedPropertyCollection(unmarshalJSON(`{ "value":[
				{
					"another_property": true,
					"a_property": "\"value\" of 'second' item"
				}
			]}`))
			Expect(err).To(BeNil())

			guid, ok := existingCollection.findGUIDForIEquivalentlItem(updatedCollection[0])
			Expect(ok).To(BeTrue())
			Expect(guid).To(Equal("28bab1d3-4a4b-48d5-8dac-two"))
		})
	})
})
