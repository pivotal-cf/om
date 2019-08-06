package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedProducts", func() {
	var (
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			Client: client,
		})
	})

	Describe("Stage", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
				}
				switch req.URL.Path {
				case "/api/v0/staged/products":
					if req.Method == "GET" {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
						}
					}
				}
				return resp, nil
			}
		})

		It("makes a request to stage the product to the Ops Manager", func() {
			err := service.Stage(api.StageProductInput{
				ProductName:    "some-product",
				ProductVersion: "some-version",
			}, "")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			By("checking for already staged products")
			checkStReq := client.DoArgsForCall(1)
			Expect(checkStReq.URL.Path).To(Equal("/api/v0/staged/products"))

			By("posting to the staged products endpoint with the product name and version")
			stReq := client.DoArgsForCall(1)
			Expect(stReq.URL.Path).To(Equal("/api/v0/staged/products"))
			Expect(stReq.Method).To(Equal("POST"))
			stReqBody, err := ioutil.ReadAll(stReq.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(stReqBody).To(MatchJSON(`{
							"name": "some-product",
							"product_version": "some-version"
			}`))
		})

		When("the same type of product is already deployed", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
					}
					switch req.URL.Path {
					case "/api/v0/staged/products":
						if req.Method == "GET" {
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
							}
						}
					}

					return resp, nil
				}
			})

			It("makes a request to stage the product to the Ops Manager", func() {
				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.1.0",
				}, "some-deployed-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(2))

				By("posting to the staged products endpoint with the product name and version")
				stReq := client.DoArgsForCall(1)
				Expect(stReq.URL.Path).To(Equal("/api/v0/staged/products/some-deployed-guid"))
				Expect(stReq.Method).To(Equal("PUT"))
				stReqBody, err := ioutil.ReadAll(stReq.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(stReqBody).To(MatchJSON(`{
					"to_version": "1.1.0"
				}`))
			})

		})

		When("the same type of product is already staged", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
					}
					switch req.URL.Path {
					case "/api/v0/staged/products":
						if req.Method == "GET" {
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body: ioutil.NopCloser(bytes.NewBufferString(`[
								{
									"type":"some-product",
									"guid": "some-staged-guid"
								},
								{
									"type":"some-other-product",
									"guid": "some-other-staged-guid"
								}]`)),
							}
						}
					}
					return resp, nil
				}
			})

			It("makes a request to stage the product to the Ops Manager", func() {
				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.1.0",
				}, "")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(2))

				By("checking for already staged products")
				depReq := client.DoArgsForCall(0)
				Expect(depReq.URL.Path).To(Equal("/api/v0/staged/products"))

				By("posting to the staged products endpoint with the product name and version")
				stReq := client.DoArgsForCall(1)
				Expect(stReq.URL.Path).To(Equal("/api/v0/staged/products/some-staged-guid"))
				Expect(stReq.Method).To(Equal("PUT"))
				stReqBody, err := ioutil.ReadAll(stReq.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(stReqBody).To(MatchJSON(`{
					"to_version": "1.1.0"
				}`))
			})

		})

		When("an error occurs", func() {
			When("a GET to the staged products endpoint returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						resp := &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
						}
						if req.URL.Path == "/api/v0/staged/products" && req.Method == "GET" {
							return nil, fmt.Errorf("some error")
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err.Error()).To(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products: some error"))
				})
			})

			When("a POST/PUT to the staged products endpoint returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						if req.Method == "GET" {
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
							}
						}
						if req.Method == "POST" && req.URL.Path == "/api/v0/staged/products" {
							return nil, fmt.Errorf("some error")
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err).To(MatchError("could not make POST api request to staged products endpoint: some error"))
				})
			})

			When("a POST/PUT to the staged products endpoint returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						if req.Method == "GET" {
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
							}
						}
						if req.URL.Path == "/api/v0/staged/products" && req.Method == "POST" {
							return &http.Response{
								StatusCode: http.StatusInternalServerError,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
							}, nil
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("GetStagedProductJobMaxInFlight", func() {
		It("makes a requests to retrieve max in flight for all jobs", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var responseBody = ioutil.NopCloser(strings.NewReader(`{
				"max_in_flight": {
					"some-third-guid": 1,
					"some-other-guid": 1,
					"some-job-guid": "20%",
					"some-fourth-guid": "default"
				}}`))

				return &http.Response{StatusCode: http.StatusOK, Body: responseBody}, nil
			}
			jobsWithMaxInFlight, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(jobsWithMaxInFlight).To(Equal(map[string]interface{}{
				"some-job-guid":    "20%",
				"some-other-guid":  1.00,
				"some-third-guid":  1.00,
				"some-fourth-guid": "default",
			}))

			Expect(client.DoCallCount()).To(Equal(1))
			request := client.DoArgsForCall(0)
			Expect(request.URL.Path).To(Equal("/api/v0/staged/products/product-type1-guid/max_in_flight"))
		})

		When("JSON response body is not valid JSON", func() {
			It("returns an error", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var responseBody = ioutil.NopCloser(strings.NewReader(`[invalidJSON}`))

					return &http.Response{StatusCode: http.StatusOK, Body: responseBody}, nil
				}
				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})

		When("the response is not 200 OK", func() {
			It("returns an error", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusNotFound}, nil
				}
				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})

		When("the request cannot be made", func() {
			It("returns an error", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("some error")
				}
				_, err := service.GetStagedProductJobMaxInFlight("product-type1-guid")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateStagedProductJobMaxInFlight", func() {
		It("makes a requests to set max in flight for all jobs", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {

				return &http.Response{StatusCode: http.StatusOK}, nil
			}
			err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
				"some-job-guid":    "20%",
				"some-other-guid":  1,
				"some-third-guid":  "1",
				"some-fourth-guid": "default",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))
			request := client.DoArgsForCall(0)
			Expect(request.URL.Path).To(Equal("/api/v0/staged/products/product-type1-guid/max_in_flight"))
			requestBody, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(requestBody).To(MatchYAML(`{
				max_in_flight: {
					some-third-guid: 1,
					some-other-guid: 1,
					some-job-guid: "20%",
					some-fourth-guid: default
			}}`))
		})

		When("no jobs are passed", func() {
			It("does nothing", func() {
				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(0))
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
				Expect(err.Error()).To(ContainSubstring(`invalid max_in_flight value provided for job 'the-guid-with-the-invalid-value': 'maximum'
valid options configurations include percentages ('50%'), counts ('2'), and 'default'`))

				Expect(client.DoCallCount()).To(Equal(0))
			})
		})

		When("the response does not return a 200 OK", func() {
			It("returns an error", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusExpectationFailed}, nil
				}
				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
					"some-job-guid":   "20%",
					"some-other-guid": 1,
				})
				Expect(err).To(HaveOccurred())
			})
		})

		When("creating the request fails", func() {
			It("returns an error", func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("some error")
				}
				err := service.UpdateStagedProductJobMaxInFlight("product-type1-guid", map[string]interface{}{
					"some-job-guid":   "20%",
					"some-other-guid": 1,
				})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteStagedProduct", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response

				switch req.URL.Path {
				case "/api/v0/staged/products":
					if req.Method == "GET" {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`[{"guid":"some-product-guid","type":"some-product"}]`)),
						}
					}
				case "/api/v0/staged/products/some-product-guid":
					if req.Method == "DELETE" {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`{"component": {"guid": "some-product-guid"}}`)),
						}
					}
				}
				return resp, nil
			}
		})

		It("makes a request to unstage the product from the Ops Manager", func() {
			err := service.DeleteStagedProduct(api.UnstageProductInput{
				ProductName: "some-product",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(2))

			By("checking for already staged products")
			checkStReq := client.DoArgsForCall(0)
			Expect(checkStReq.URL.Path).To(Equal("/api/v0/staged/products"))

			By("deleting from the set of staged products")
			deleteReq := client.DoArgsForCall(1)
			Expect(deleteReq.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid"))
			Expect(deleteReq.Method).To(Equal("DELETE"))
			_, err = ioutil.ReadAll(deleteReq.Body)
			Expect(err).NotTo(HaveOccurred())
		})

		When("the product is not staged", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response
					if req.URL.Path == "/api/v0/staged/products" && req.Method == "GET" {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`[{"guid":"some-other-product-guid","type":"some-other-product"}]`)),
						}
					}
					return resp, nil
				}
			})

			It("returns an error", func() {
				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("product is not staged: some-product"))
			})
		})

		When("a GET to the staged products endpoint returns an error", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
					}
					if req.URL.Path == "/api/v0/staged/products" && req.Method == "GET" {
						return nil, fmt.Errorf("some error")
					}
					return resp, nil
				}
			})

			It("returns an error", func() {
				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products: some error"))
			})
		})

		When("a DELETE to the staged products endpoint returns an error", func() {
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response

					switch req.URL.Path {
					case "/api/v0/staged/products":
						if req.Method == "GET" {
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`[{"guid":"some-product-guid","type":"some-product"}]`)),
							}
						}
					case "/api/v0/staged/products/some-product-guid":
						if req.Method == "DELETE" {
							return nil, fmt.Errorf("some error")
						}
					}
					return resp, nil
				}
			})

			It("returns an error", func() {
				err := service.DeleteStagedProduct(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("could not send api request to DELETE /api/v0/staged/products/some-product-guid: some error"))
			})
		})
	})

	Describe("ListStagedProducts", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
				}
				switch req.URL.Path {
				case "/api/v0/staged/products":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`[{
							"guid":"some-product-guid",
							"type":"some-type"
						},
						{
							"guid":"some-other-product-guid",
							"type":"some-other-type"
						}]`)),
					}
				}
				return resp, nil
			}
		})

		It("retrieves a list of staged products from the Ops Manager", func() {
			output, err := service.ListStagedProducts()
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(Equal(api.StagedProductsOutput{
				Products: []api.StagedProduct{
					{
						GUID: "some-product-guid",
						Type: "some-type",
					},
					{
						GUID: "some-other-product-guid",
						Type: "some-other-type",
					},
				},
			}))

			Expect(client.DoCallCount()).To(Equal(1))

			By("checking for staged products")
			avReq := client.DoArgsForCall(0)
			Expect(avReq.URL.Path).To(Equal("/api/v0/staged/products"))
		})

		Context("failure cases", func() {
			When("the request fails", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))
				})

				It("returns an error", func() {
					_, err := service.ListStagedProducts()
					Expect(err.Error()).To(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products: nope"))
				})
			})

			When("the server returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)
				})

				It("returns an error", func() {
					_, err := service.ListStagedProducts()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			When("the server returns invalid JSON", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString("%%")),
					}, nil)
				})

				It("returns an error", func() {
					_, err := service.ListStagedProducts()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged products response:")))
				})
			})
		})
	})

	Describe("UpdateStagedProductProperties", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/products/some-product-guid/properties":
					if req.Method == "GET" {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
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
      "value": [
        {
          "guid": {
            "type": "uuid",
            "configurable": false,
            "credential": false,
            "value": "28bab1d3-4a4b-48d5-8dac-796adf078100",
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
        }
      ],
      "optional": false
    },
  }
}
`)),
						}
					} else {
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
						}
					}
				default:
					Fail(fmt.Sprintf("unexpected request to '%s'", req.URL.Path))
				}
				return resp, nil
			}
		})

		It("configures the properties for the given staged product in the Ops Manager", func() {
			err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
				GUID: "some-product-guid",
				Properties: `{
					"key": "value"
				}`,
			})
			Expect(err).NotTo(HaveOccurred())

			By("configuring the product properties")
			Expect(client.DoCallCount()).To(Equal(2))
			req := client.DoArgsForCall(1)
			Expect(req.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/properties"))
			Expect(req.Method).To(Equal("PUT"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			reqBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(reqBody).To(MatchJSON(`{
				"properties": {
					"key": "value"
				}
			}`))
		})
		Context("configure product contains collection", func() {
			It("adds the guid for elements that exist", func() {
				err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
					GUID: "some-product-guid",
					Properties: `{
					"key": "value",
					"some_collection": {
						"value": [
							{
								"name": "the_name",
								"some_property": "property_value"
							}
						]
					}
				}`,
				})
				Expect(err).NotTo(HaveOccurred())

				By("configuring the product properties")
				Expect(client.DoCallCount()).To(Equal(2))
				req := client.DoArgsForCall(1)
				Expect(req.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/properties"))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				reqBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBody).To(MatchJSON(`{
				"properties": {
					"key": "value",
					"some_collection": {
						"value": [
							{
								"name": "the_name",
								"some_property": "property_value",
								"guid": "28bab1d3-4a4b-48d5-8dac-796adf078100"
							}
						]
					}
				}
			}`))
			})
			It("no guid added", func() {
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
				Expect(err).NotTo(HaveOccurred())

				By("configuring the product properties")
				Expect(client.DoCallCount()).To(Equal(2))
				req := client.DoArgsForCall(1)
				Expect(req.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/properties"))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				reqBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBody).To(MatchJSON(`{
				"properties": {
					"key": "value",
					"some_other_collection": {
						"value": [
							{
								"name": "other_name",
								"some_property": "property_value"
							}
						]
					}
				}
			}`))
			})
			It("does not contain a name in the collection element", func() {
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
				Expect(err).NotTo(HaveOccurred())

				By("configuring the product properties")
				Expect(client.DoCallCount()).To(Equal(2))
				req := client.DoArgsForCall(1)
				Expect(req.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/properties"))
				Expect(req.Method).To(Equal("PUT"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				reqBody, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqBody).To(MatchJSON(`{
					"properties": {
						"key": "value",
						"some_collection": {
							"value": [
								{
									"some_property": "property_value"
								}
							]
						}
					}
				}`))

			})
		})

		Context("failure cases", func() {
			When("the request fails", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))
				})

				It("returns an error", func() {
					err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
						GUID:       "foo",
						Properties: `{}`,
					})
					Expect(err).To(MatchError("could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/foo/properties: nope"))
				})
			})

			When("the server returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)
				})

				It("returns an error", func() {
					err := service.UpdateStagedProductProperties(api.UpdateStagedProductPropertiesInput{
						GUID:       "foo",
						Properties: `{}`,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("UpdateStagedProductNetworksAndAZs", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/products/some-product-guid/properties":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
					}
				case "/api/v0/staged/products/some-product-guid/networks_and_azs":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
					}
				}
				return resp, nil
			}
		})

		It("configures the networks for the given staged product in the Ops Manager", func() {
			err := service.UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput{
				GUID: "some-product-guid",
				NetworksAndAZs: `{
					"key": "value"
				}`,
			})
			Expect(err).NotTo(HaveOccurred())

			By("configuring the product properties")
			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)
			Expect(req.URL.Path).To(Equal("/api/v0/staged/products/some-product-guid/networks_and_azs"))
			Expect(req.Method).To(Equal("PUT"))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

			reqBody, err := ioutil.ReadAll(req.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(reqBody).To(MatchJSON(`{
				"networks_and_azs": {
					"key": "value"
				}
			}`))
		})

		Context("failure cases", func() {
			When("the request fails", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))
				})

				It("returns an error", func() {
					err := service.UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput{
						GUID:           "foo",
						NetworksAndAZs: `{}`,
					})
					Expect(err).To(MatchError("could not make api request to staged product networks_and_azs endpoint: could not send api request to PUT /api/v0/staged/products/foo/networks_and_azs: nope"))
				})
			})

			When("the server returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)
				})

				It("returns an error", func() {
					err := service.UpdateStagedProductNetworksAndAZs(api.UpdateStagedProductNetworksAndAZsInput{
						GUID:           "foo",
						NetworksAndAZs: `{}`,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("GetStagedProductManifest", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/products/some-product-guid/manifest":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
							"manifest": {
								"key-1": {
									"key-2": "value-1"
								},
								"key-3": "value-2",
								"key-4": 2147483648
							}
						}`)),
					}
				}
				return resp, nil
			}
		})

		It("returns the manifest for a product", func() {
			manifest, err := service.GetStagedProductManifest("some-product-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(MatchYAML(`---
key-1:
  key-2: value-1
key-3: value-2
key-4: 2147483648
`))
		})

		Context("failure cases", func() {
			When("the request object is invalid", func() {
				It("returns an error", func() {
					_, err := service.GetStagedProductManifest("invalid-guid-%%%")
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			When("the client request fails", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))

					_, err := service.GetStagedProductManifest("some-product-guid")
					Expect(err).To(MatchError("could not make api request to staged products manifest endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/manifest: nope"))
				})
			})

			When("the server returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)

					_, err := service.GetStagedProductManifest("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			When("the returned JSON is invalid", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString("---some-malformed-json")),
					}, nil)

					_, err := service.GetStagedProductManifest("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedProductProperties", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/products/some-product-guid/properties":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
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
						}`)),
					}
				}
				return resp, nil
			}
		})

		It("returns the configuration for a product", func() {
			config, err := service.GetStagedProductProperties("some-product-guid")
			Expect(err).NotTo(HaveOccurred())

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

		Context("failure cases", func() {
			When("the properties request returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/products/some-product-guid/properties":
							return &http.Response{}, errors.New("some-error")
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedProductProperties("some-product-guid")
					Expect(err).To(MatchError(`could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/properties: some-error`))
				})
			})

			When("the properties request returns a non 200 error code", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/products/some-product-guid/properties":
							return &http.Response{
								StatusCode: http.StatusTeapot,
								Body:       ioutil.NopCloser(bytes.NewBufferString("")),
							}, nil
						}
						return resp, nil
					}
				})
				It("returns an error", func() {
					_, err := service.GetStagedProductProperties("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			When("the server returns invalid json", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/products/some-product-guid/properties":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{{{`)),
							}
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedProductProperties("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})

	Describe("GetStagedProductNetworksAndAZs", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				switch req.URL.Path {
				case "/api/v0/staged/products/some-product-guid/networks_and_azs":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
							"networks_and_azs": {
						  	"singleton_availability_zone": {
                  "name": "az-one"
                },
                "other_availability_zones": [
                  { "name": "az-two" },
                  { "name": "az-three" }
                ],
                "network": {
                  "name": "network-one"
                }
						  }
						}`)),
					}
				}
				return resp, nil
			}
		})

		It("returns the networks + azs for a product", func() {
			config, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
			Expect(err).NotTo(HaveOccurred())
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
			BeforeEach(func() {
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response
					var err error
					switch req.URL.Path {
					case "/api/v0/staged/products/some-product-guid/networks_and_azs":
						resp = &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       ioutil.NopCloser(nil),
						}
					}
					return resp, err
				}
			})

			It("returns an empty payload without error", func() {
				config, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(map[string]interface{}(nil)))
			})
		})

		Context("failure cases", func() {
			When("the networks_and_azs request returns an error", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/products/some-product-guid/networks_and_azs":
							return &http.Response{}, errors.New("some-error")
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
					Expect(err).To(MatchError(`could not make api request to staged product properties endpoint: could not send api request to GET /api/v0/staged/products/some-product-guid/networks_and_azs: some-error`))
				})
			})

			When("the server returns invalid json", func() {
				BeforeEach(func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						switch req.URL.Path {
						case "/api/v0/staged/products/some-product-guid/networks_and_azs":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{{{`)),
							}
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					_, err := service.GetStagedProductNetworksAndAZs("some-product-guid")
					Expect(err).To(MatchError(ContainSubstring("could not parse json")))
				})
			})
		})
	})
})
