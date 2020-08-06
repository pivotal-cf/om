package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
)

var _ = Describe("StagedProducts", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Describe("Stage", func() {
		It("makes a request to stage the product to the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v0/staged/products"),
					ghttp.VerifyJSON(`{
						"name": "some-product",
						"product_version": "some-version"
					}`),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			err := service.Stage(api.StageProductInput{
				ProductName:    "some-product",
				ProductVersion: "some-version",
			}, "")
			Expect(err).ToNot(HaveOccurred())
		})

		When("the same type of product is already deployed", func() {
			It("makes a request to stage the product to the Ops Manager", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-deployed-guid"),
						ghttp.VerifyJSON(`{
							"to_version": "1.1.0"
						}`),
						ghttp.RespondWith(http.StatusOK, ``),
					),
				)

				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.1.0",
				}, "some-deployed-guid")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the same type of product is already staged", func() {
			It("makes a request to stage the product to the Ops Manager", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"type":"some-product",
							"guid": "some-staged-guid"
						}, {
							"type":"some-other-product",
							"guid": "some-other-staged-guid"
						}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-staged-guid"),
						ghttp.VerifyJSON(`{
							"to_version": "1.1.0"
						}`),
						ghttp.RespondWith(http.StatusOK, ``),
					),
				)

				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.1.0",
				}, "")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("a GET to the staged products endpoint returns an error", func() {
			It("returns an error", func() {
				client.Close()

				err := service.Stage(api.StageProductInput{
					ProductName:    "foo",
					ProductVersion: "bar",
				}, "")
				Expect(err).To(MatchError(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products")))
			})
		})

		When("a POST/PUT to the staged products endpoint returns an error", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/staged/products"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							client.CloseClientConnections()
						}),
					),
				)

				err := service.Stage(api.StageProductInput{
					ProductName:    "foo",
					ProductVersion: "bar",
				}, "")
				Expect(err).To(MatchError(ContainSubstring("could not make POST api request to staged products endpoint")))
			})
		})

		When("a POST/PUT to the staged products endpoint returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.Stage(api.StageProductInput{
					ProductName:    "foo",
					ProductVersion: "bar",
				}, "")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("GetStagedProductJobMaxInFlight", func() {
		It("makes a requests to retrieve max in flight for all jobs", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/product-type1-guid/max_in_flight"),
					ghttp.RespondWith(http.StatusOK, `{
						"max_in_flight": {
							"some-third-guid": 1,
							"some-other-guid": 1,
							"some-job-guid": "20%",
							"some-fourth-guid": "default"
						}
					}`),
				),
			)

			jobsWithMaxInFlight, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
			Expect(err).ToNot(HaveOccurred())
			Expect(jobsWithMaxInFlight).To(Equal(map[string]interface{}{
				"some-job-guid":    "20%",
				"some-other-guid":  1.00,
				"some-third-guid":  1.00,
				"some-fourth-guid": "default",
			}))
		})

		When("JSON response body is not valid JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/product-type1-guid/max_in_flight"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})

		When("the response is not 200 OK", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/product-type1-guid/max_in_flight"),
						ghttp.RespondWith(http.StatusTeapot, ``),
					),
				)

				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})

		When("the request cannot be made", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateStagedProductJobMaxInFlight", func() {
		It("makes a requests to set max in flight for all jobs", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/product-type1-guid/max_in_flight"),
					ghttp.VerifyJSON(`{
						"max_in_flight": {
							"some-third-guid": 1,
							"some-other-guid": 1,
							"some-job-guid": "20%",
							"some-fourth-guid": "default"
						}
					}`),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
				"some-job-guid":    "20%",
				"some-other-guid":  1,
				"some-third-guid":  "1",
				"some-fourth-guid": "default",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("no jobs are passed", func() {
			It("does nothing", func() {
				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{})
				Expect(err).ToNot(HaveOccurred())

				Expect(len(client.ReceivedRequests())).To(Equal(0))
			})
		})

		When("an invalid value is provided for max_in_flight", func() {
			It("prints an error indicating the valid formats", func() {
				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
					"some-job-guid":                   "20%",
					"some-other-guid":                 1,
					"some-third-guid":                 "1",
					"some-fourth-guid":                "default",
					"the-guid-with-the-invalid-value": "maximum",
				})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring(`invalid max_in_flight value provided for job 'the-guid-with-the-invalid-value': 'maximum'
valid options configurations include percentages ('50%'), counts ('2'), and 'default'`)))

				Expect(len(client.ReceivedRequests())).To(Equal(0))
			})
		})

		When("the response does not return a 200 OK", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/product-type1-guid/max_in_flight"),
						ghttp.RespondWith(http.StatusExpectationFailed, ``),
					),
				)

				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
					"some-job-guid":   "20%",
					"some-other-guid": 1,
				})
				Expect(err).To(HaveOccurred())
			})
		})

		When("creating the request fails", func() {
			It("returns an error", func() {
				client.Close()

				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
					"some-job-guid":   "20%",
					"some-other-guid": 1,
				})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteStagedProduct", func() {
		It("makes a request to unstage the product from the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{
						"guid": "some-product-guid",
						"type": "some-product"
					}]`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/staged/products/some-product-guid"),
					ghttp.RespondWith(http.StatusOK, `{
						"component": {
							"guid": "some-product-guid"
						}
					}`),
				),
			)

			err := service.DeleteStagedProduct(api.UnstageProductInput{
				ProductName: "some-product",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		When("the product is not staged", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"guid": "some-product-guid",
							"type": "some-product"
						}]`),
					),
				)

				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-other-product",
				})
				Expect(err).To(MatchError("product is not staged: some-other-product"))
			})
		})

		When("a GET to the staged products endpoint returns an error", func() {
			It("returns an error", func() {
				client.Close()

				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products")))
			})
		})

		When("a DELETE to the staged products endpoint returns an error", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"guid": "some-product-guid",
							"type": "some-product"
						}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/v0/staged/products/some-product-guid"),
						http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
							client.CloseClientConnections()
						}),
					),
				)

				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError(ContainSubstring("could not send api request to DELETE /api/v0/staged/products/some-product-guid")))
			})
		})
	})

	Describe("ListStagedProducts", func() {
		It("retrieves a list of staged products from the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
					ghttp.RespondWith(http.StatusOK, `[{
							"guid":"some-product-guid",
							"type":"some-type"
						}, {
							"guid":"some-other-product-guid",
							"type":"some-other-type"
						}]`),
				),
			)

			output, err := service.ListStagedProducts()
			Expect(err).ToNot(HaveOccurred())

			Expect(output).To(Equal(api.StagedProductsOutput{
				Products: []api.StagedProduct{{
					GUID: "some-product-guid",
					Type: "some-type",
				}, {
					GUID: "some-other-product-guid",
					Type: "some-other-type",
				}},
			}))
		})

		When("the request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.ListStagedProducts()
				Expect(err).To(MatchError(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusTeapot, ``),
					),
				)

				_, err := service.ListStagedProducts()
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the server returns invalid JSON", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.ListStagedProducts()
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged products response:")))
			})
		})
	})

	Describe("UpdateStagedProductProperties", func() {
		BeforeEach(func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties"),
					ghttp.RespondWith(http.StatusOK, `{
						"properties": {
							"sample": {
								"type": "string",
								"configurable": false,
								"credential": false,
								"value": "account",
								"optional": false
							},
							"some_collection": {
								"type": "collection",
								"configurable": true,
								"credential": false,
								"value": [{
									"guid": {
										"type": "uuid",
										"configurable": false,
										"credential": false,
										"value": "28bab1d3-4a4b-48d5-8dac-796adf078100",
										"optional": false
									},
									"label": {
										"type": "string",
										"configurable": true,
										"credential": false,
										"value": "the_label",
										"optional": false
									},
									"some_property": {
										"type": "boolean",
										"configurable": true,
										"credential": false,
										"value": true,
										"optional": false
									}
								}],
								"optional": false
							},
							"collection_with_name": {
								"type": "collection",
								"configurable": true,
								"credential": false,
								"value": [{
									"guid": {
										"type": "uuid",
										"configurable": false,
										"credential": false,
										"value": "28bab1d3-4a4b-48d5-8dac-with-name",
										"optional": false
									},
									"name": {
										"type": "string",
										"configurable": true,
										"credential": false,
										"value": "the_name",
										"optional": false
									},
									"some_property": {
										"type": "boolean",
										"configurable": true,
										"credential": false,
										"value": true,
										"optional": false
									}
								}],
								"optional": false
							},
							"collection_with_key": {
								"type": "collection",
								"configurable": true,
								"credential": false,
								"value": [{
									"guid": {
										"type": "uuid",
										"configurable": false,
										"credential": false,
										"value": "28bab1d3-4a4b-48d5-8dac-with-key",
										"optional": false
									},
									"key": {
										"type": "string",
										"configurable": true,
										"credential": false,
										"value": "the_key_value",
										"optional": false
									},
									"some_property": {
										"type": "boolean",
										"configurable": true,
										"credential": false,
										"value": true,
										"optional": false
									}
								}],
								"optional": false
							},
							"collection_with_logical_key_ending_in_name": {
								"type": "collection",
								"configurable": true,
								"credential": false,
								"value": [{
									"guid": {
										"type": "uuid",
										"configurable": false,
										"credential": false,
										"value": "28bab1d3-4a4b-48d5-8dac-ending-in-name",
										"optional": false
									},
									"sqlServerName": {
										"type": "string",
										"configurable": true,
										"credential": false,
										"value": "the_sql_server",
										"optional": false
									},
									"some_property": {
										"type": "boolean",
										"configurable": true,
										"credential": false,
										"value": true,
										"optional": false
									}
								}],
								"optional": false
							}
						}
					}`),
				),
			)
		})

		It("configures the properties for the given staged product in the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
						"properties": {
							"key": "value"
						}
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
				GUID: "some-product-guid",
				Properties: `{
					"key": "value"
				}`,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("configure product contains collection", func() {
			It("adds the guid for elements that exist and haven't changed, but don't have a logical key field", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"some_collection": {
									"value": [{
										"label": "the_label",
										"some_property": true,
										"guid": "28bab1d3-4a4b-48d5-8dac-796adf078100"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
						"key": "value",
						"some_collection": {
							"value": [
								{
									"some_property": true,
									"label": "the_label"
								}
							]
						}
					}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			It("adds the guid for elements that have a name property", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"collection_with_name": {
									"value": [{
										"name": "the_name",
										"some_property": "new_property_value",
										"guid": "28bab1d3-4a4b-48d5-8dac-with-name"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
						"key": "value",
						"collection_with_name": {
							"value": [
								{
									"name": "the_name",
									"some_property": "new_property_value"
								}
							]
						}
					}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			It("adds the guid for elements that have a key property", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"collection_with_key": {
									"value": [{
										"key": "the_key_value",
										"some_property": "new_property_value",
										"guid": "28bab1d3-4a4b-48d5-8dac-with-key"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
						"key": "value",
						"collection_with_key": {
							"value": [
								{
									"key": "the_key_value",
									"some_property": "new_property_value"
								}
							]
						}
					}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			It("adds the guid for elements that have logical key property ending in 'name'", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"collection_with_logical_key_ending_in_name": {
									"value": [{
										"sqlServerName": "the_sql_server",
										"some_property": "new_property_value",
										"guid": "28bab1d3-4a4b-48d5-8dac-ending-in-name"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
						"key": "value",
						"collection_with_logical_key_ending_in_name": {
							"value": [
								{
									"sqlServerName": "the_sql_server",
									"some_property": "new_property_value"
								}
							]
						}
					}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})
			It("no guid added", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"some_other_collection": {
									"value": [{
										"name": "other_name",
										"some_property": "property_value"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
						"key": "value",
						"some_other_collection": {
							"value": [
								{
									"name": "other_name",
									"some_property": "property_value"
								}
							]
						}
					}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not contain a name in the collection element", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.VerifyContentType("application/json"),
						ghttp.VerifyJSON(`{
							"properties": {
								"key": "value",
								"some_collection": {
									"value": [{
										"some_property": "property_value"
									}]
								}
							}
						}`),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
							"key": "value",
							"some_collection": {
								"value": [
									{
										"some_property": "property_value"
									}
								]
							}
						}`,
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("the request fails", func() {
			It("returns an error", func() {
				client.Close()

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID:       "foo",
					Properties: `{}`,
				})
				Expect(err).To(MatchError(ContainSubstring("could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/foo/properties")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns an error", func() {
				client.Reset()
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/foo/properties"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID:       "foo",
					Properties: `{}`,
				})
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})
	})

	Describe("GetStagedProductSyslogConfiguration", func() {
		//BeforeEach(func() {
		//	client.DoStub = func(req *http.Request) (*http.Response, error) {
		//		var resp *http.Response
		//		if strings.Contains(req.URL.Path, "some-product-guid") {
		//			resp = &http.Response{
		//				StatusCode: http.StatusOK,
		//				Body: ioutil.NopCloser(bytes.NewBufferString(`{
		//						"syslog_configuration": {
		//							"enabled": true,
		//							"address": "example.com"
		//						}
		//					}`)),
		//			}
		//		} else if strings.Contains(req.URL.Path, "missing-syslog-config") {
		//			resp = &http.Response{
		//				StatusCode: http.StatusUnprocessableEntity,
		//				Body: ioutil.NopCloser(bytes.NewBufferString(`{
		//						"errors": {
		//						  "syslog": ["This product does not support the Ops Manager consistent syslog configuration feature. If the product supports custom syslog configuration, those properties can be set via the /api/v0/staged/products/:product_guid/properties endpoint."]
		//						}
		//					}`)),
		//			}
		//		} else if strings.Contains(req.URL.Path, "bad-response-code") {
		//			resp = &http.Response{
		//				StatusCode: http.StatusBadRequest,
		//				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
		//			}
		//		} else {
		//			resp = &http.Response{
		//				StatusCode: http.StatusOK,
		//				Body: ioutil.NopCloser(bytes.NewBufferString(`{
		//						invalid-json
		//					}`)),
		//			}
		//		}
		//
		//		return resp, nil
		//	}
		//})

		When("syslog configuration is configured for the specified product", func() {
			It("returns the syslog configuration ", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/syslog_configuration"),
						ghttp.VerifyContentType("application/json"),
						ghttp.RespondWith(http.StatusOK, `{
							"syslog_configuration": {
								"enabled": true,
								"address": "example.com"
							}
						}`),
					),
				)

				syslogConfig, err := service.GetStagedProductSyslogConfiguration("some-product-guid")
				Expect(err).ToNot(HaveOccurred())

				expectedResult := make(map[string]interface{})
				expectedResult["enabled"] = true
				expectedResult["address"] = "example.com"

				Expect(syslogConfig).To(Equal(expectedResult))
			})
		})

		When("the request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetStagedProductSyslogConfiguration("some-product-guid")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("could not make api request to staged product syslog_configuration endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/syslog_configuration")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns nil when the status code is unprocessable (422)", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/missing-syslog-config/syslog_configuration"),
						ghttp.RespondWith(http.StatusUnprocessableEntity, `{}`),
					),
				)

				syslogConfig, err := service.GetStagedProductSyslogConfiguration("missing-syslog-config")
				Expect(err).ToNot(HaveOccurred())

				Expect(syslogConfig).To(BeNil())
			})

			It("returns an error for any other status code that is non-200", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/bad-response-code/syslog_configuration"),
						ghttp.RespondWith(http.StatusTeapot, `{}`),
					),
				)

				_, err := service.GetStagedProductSyslogConfiguration("bad-response-code")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the response body cannot be decoded", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/bad-json/syslog_configuration"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetStagedProductSyslogConfiguration("bad-json")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged product syslog_configuration response")))
			})
		})
	})

	Describe("UpdateStagedProductSyslogConfiguration", func() {
		It("configures the syslog for the given staged product in the Ops Manager", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/staged/products/some-product-guid/syslog_configuration"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{
						"syslog_configuration": {
							"key": "value"
						}
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.UpdateSyslogConfiguration(api.UpdateSyslogConfigurationInput{
				GUID: "some-product-guid",
				SyslogConfiguration: `{
						"key": "value"
					}`,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("failure cases", func() {
			When("the request fails", func() {
				It("returns an error", func() {
					client.Close()

					err := service.UpdateSyslogConfiguration(api.UpdateSyslogConfigurationInput{
						GUID:                "foo",
						SyslogConfiguration: `{}`,
					})
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("could not make api request to staged product syslog_configuration endpoint: could not send api request to PUT /api/v0/staged/products/foo/syslog_configuration")))
				})
			})

			When("the server returns a non-200 status code", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/api/v0/staged/products/foo/syslog_configuration"),
							ghttp.RespondWith(http.StatusTeapot, `{}`),
						),
					)

					err := service.UpdateSyslogConfiguration(api.UpdateSyslogConfigurationInput{
						GUID:                "foo",
						SyslogConfiguration: `{}`,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("GetStagedProductManifest", func() {
		It("returns the manifest for a product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/manifest"),
					ghttp.RespondWith(http.StatusOK, `{
						"manifest": {
							"key-1": {
								"key-2": "value-1"
							},
							"key-3": "value-2",
							"key-4": 2147483648
						}
					}`),
				),
			)

			manifest, err := service.GetStagedProductManifest("some-product-guid")
			Expect(err).ToNot(HaveOccurred())
			Expect(manifest).To(MatchYAML(`---
key-1:
  key-2: value-1
key-3: value-2
key-4: 2147483648
`))
		})

		When("the request object is invalid", func() {
			It("returns an error", func() {
				_, err := service.GetStagedProductManifest("invalid-guid-%%%")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		When("the client request fails", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetStagedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not make api request to staged products manifest endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/manifest")))
			})
		})

		When("the server returns a non-200 status code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/manifest"),
						ghttp.RespondWith(http.StatusTeapot, ``),
					),
				)

				_, err := service.GetStagedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the returned JSON is invalid", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/manifest"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetStagedProductManifest("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not parse json")))
			})
		})
	})

	Describe("GetStagedProductProperties", func() {
		It("returns the configuration for a product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties"),
					ghttp.RespondWith(http.StatusOK, `{
						"properties": {
							".properties.some-configurable-property": {
								"value": "some-value",
								"configurable": true
							},
							".properties.some-non-configurable-property": {
								"value": "some-value",
								"configurable": false
							},
							".properties.some-secret-property": {
								"value": {
									"some-secret-type": "***"
								},
								"configurable": true,
								"credential": true
							},
							".properties.some-selector-property": {
								"value": "Plan 1",
								"selected_option": "xGB",
								"configurable": true
							},
							".properties.some-property-with-a-large-number-value": {
								"value": 2147483648,
								"configurable": true
							}
						}
					}`),
				),
			)

			config, err := service.GetStagedProductProperties("some-product-guid")
			Expect(err).ToNot(HaveOccurred())

			Expect(config).To(HaveKeyWithValue(".properties.some-configurable-property", api.ResponseProperty{
				Value:        "some-value",
				Configurable: true,
				IsCredential: false,
			}))
			Expect(config).To(HaveKeyWithValue(".properties.some-non-configurable-property", api.ResponseProperty{
				Value:        "some-value",
				Configurable: false,
				IsCredential: false,
			}))
			Expect(config).To(HaveKeyWithValue(".properties.some-secret-property", api.ResponseProperty{
				Value: map[interface{}]interface{}{
					"some-secret-type": "***",
				},
				Configurable: true,
				IsCredential: true,
			}))
			Expect(config).To(HaveKeyWithValue(".properties.some-property-with-a-large-number-value", api.ResponseProperty{
				Value:        2147483648,
				Configurable: true,
				IsCredential: false,
			}))
			Expect(config).To(HaveKeyWithValue(".properties.some-selector-property", api.ResponseProperty{
				Value:          "Plan 1",
				SelectedOption: "xGB",
				Configurable:   true,
				IsCredential:   false,
			}))
		})

		When("the properties request returns an error", func() {
			It("returns an error", func() {
				client.Close()

				_, err := service.GetStagedProductProperties("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring(`could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/properties`)))
			})
		})

		When("the properties request returns a non 200 error code", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.RespondWith(http.StatusTeapot, ``),
					),
				)

				_, err := service.GetStagedProductProperties("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
			})
		})

		When("the server returns invalid json", func() {
			It("returns an error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/properties"),
						ghttp.RespondWith(http.StatusOK, `invalid-json`),
					),
				)

				_, err := service.GetStagedProductProperties("some-product-guid")
				Expect(err).To(MatchError(ContainSubstring("could not parse json")))
			})
		})
	})

	Describe("GetStagedProductNetworksAndAZs", func() {
		It("returns the networks + azs for a product", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
					ghttp.RespondWith(http.StatusOK, `{
						"networks_and_azs": {
							"singleton_availability_zone": {
								"name": "az-one"
							},
							"other_availability_zones": [{
								"name": "az-two"
							}, {
								"name": "az-three"
							}],
							"network": {
								"name": "network-one"
							}
						}
					}`),
				),
			)

			config, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(map[string]interface{}{
				"singleton_availability_zone": map[string]interface{}{
					"name": "az-one",
				},
				"other_availability_zones": []interface{}{
					map[string]interface{}{"name": "az-two"},
					map[string]interface{}{"name": "az-three"},
				},
				"network": map[string]interface{}{
					"name": "network-one",
				},
			}))
		})

		When("there is no network + azs for the give product", func() {
			It("returns an empty payload without error", func() {
				client.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
						ghttp.RespondWith(http.StatusNotFound, ``),
					),
				)

				config, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(config).To(Equal(map[string]interface{}(nil)))
			})
		})

		Context("failure cases", func() {
			When("the networks_and_azs request returns an error", func() {
				It("returns an error", func() {
					client.Close()

					_, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring(`could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/networks_and_azs`)))
				})
			})

			When("the server returns invalid json", func() {
				It("returns an error", func() {
					client.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/staged/products/some-product-guid/networks_and_azs"),
							ghttp.RespondWith(http.StatusOK, `invalid-json`),
						),
					)

					_, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})
})
