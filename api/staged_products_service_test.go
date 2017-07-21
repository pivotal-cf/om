package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagedProductsService", func() {
	Describe("Stage", func() {
		var (
			client *fakes.HttpClient
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				resp = &http.Response{
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
			service := api.NewStagedProductsService(client)

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

		Context("when the same type of product is already deployed", func() {
			BeforeEach(func() {
				client = &fakes.HttpClient{}
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response
					resp = &http.Response{
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
				service := api.NewStagedProductsService(client)

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

		Context("when the same type of product is already staged", func() {
			BeforeEach(func() {
				client = &fakes.HttpClient{}
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response
					resp = &http.Response{
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
				service := api.NewStagedProductsService(client)

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

		Context("when an error occurs", func() {
			Context("when a GET to the staged products endpoint returns an error", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						resp = &http.Response{
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
					service := api.NewStagedProductsService(client)

					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err).To(MatchError("could not make api request to staged products endpoint: some error"))
				})
			})

			Context("when a POST/PUT to the staged products endpoint returns an error", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
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
					service := api.NewStagedProductsService(client)

					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err).To(MatchError("could not make POST api request to staged products endpoint: some error"))
				})
			})

			Context("when a POST/PUT to the staged products endpoint returns a non-200 status code", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
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
					service := api.NewStagedProductsService(client)
					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					}, "")
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("Unstage", func() {
		var (
			client *fakes.HttpClient
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
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
			service := api.NewStagedProductsService(client)

			err := service.Unstage(api.UnstageProductInput{
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

		Context("when the product is not staged", func() {
			BeforeEach(func() {
				client = &fakes.HttpClient{}

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
				service := api.NewStagedProductsService(client)

				err := service.Unstage(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("product is not staged: some-product"))
			})
		})

		Context("when a GET to the staged products endpoint returns an error", func() {
			BeforeEach(func() {
				client = &fakes.HttpClient{}
				client.DoStub = func(req *http.Request) (*http.Response, error) {
					var resp *http.Response
					resp = &http.Response{
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
				service := api.NewStagedProductsService(client)

				err := service.Unstage(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("could not make api request to staged products endpoint: some error"))
			})
		})

		Context("when a DELETE to the staged products endpoint returns an error", func() {
			BeforeEach(func() {

				client = &fakes.HttpClient{}

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
				service := api.NewStagedProductsService(client)

				err := service.Unstage(api.UnstageProductInput{
					ProductName: "some-product",
				})
				Expect(err).To(MatchError("could not make DELETE api request to staged products endpoint: some error"))
			})
		})
	})

	Describe("StagedProducts", func() {
		var (
			client *fakes.HttpClient
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				var resp *http.Response
				resp = &http.Response{
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
			service := api.NewStagedProductsService(client)

			output, err := service.StagedProducts()
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
			Context("when the request fails", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))
				})

				It("returns an error", func() {
					service := api.NewStagedProductsService(client)

					_, err := service.StagedProducts()
					Expect(err).To(MatchError("could not make api request to staged products endpoint: nope"))
				})
			})

			Context("when the server returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)
				})

				It("returns an error", func() {
					service := api.NewStagedProductsService(client)

					_, err := service.StagedProducts()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the server returns invalid JSON", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString("%%")),
					}, nil)
				})

				It("returns an error", func() {
					service := api.NewStagedProductsService(client)

					_, err := service.StagedProducts()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged products response:")))
				})
			})
		})
	})

	Describe("Configure", func() {
		var (
			client *fakes.HttpClient
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
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

		It("configures the properties for the given staged product in the Ops Manager", func() {
			service := api.NewStagedProductsService(client)

			err := service.Configure(api.ProductsConfigurationInput{
				GUID: "some-product-guid",
				Configuration: `{
					"key": "value"
				}`,
			})
			Expect(err).NotTo(HaveOccurred())

			By("configuring the product properties")
			Expect(client.DoCallCount()).To(Equal(1))
			req := client.DoArgsForCall(0)
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

		It("configures the network for the given staged product in the Ops Manager", func() {
			service := api.NewStagedProductsService(client)

			err := service.Configure(api.ProductsConfigurationInput{
				GUID: "some-product-guid",
				Network: `{
					"key": "value"
				}`,
			})
			Expect(err).NotTo(HaveOccurred())

			By("configuring the product network")
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
			Context("when the request fails", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{}, errors.New("nope"))
				})

				It("returns an error", func() {
					service := api.NewStagedProductsService(client)

					err := service.Configure(api.ProductsConfigurationInput{
						GUID:          "foo",
						Configuration: `{}`,
					})
					Expect(err).To(MatchError("could not make api request to staged product properties endpoint: nope"))
				})
			})

			Context("when the server returns a non-200 status code", func() {
				BeforeEach(func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					}, nil)
				})

				It("returns an error", func() {
					service := api.NewStagedProductsService(client)

					err := service.Configure(api.ProductsConfigurationInput{
						GUID:          "foo",
						Configuration: `{}`,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("Find", func() {
		var (
			client  *fakes.HttpClient
			service api.StagedProductsService
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			service = api.NewStagedProductsService(client)
		})

		It("Find product by product name", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewBufferString(`[
					{"installation_name":"p-bosh","guid":"some-product-id","type":"some-product-name","product_version":"1.10.0.0"},
					{"installation_name":"cf-15b22d1810a034ea3aca","guid":"cf-15b22d1810a034ea3aca","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"p-isolation-segment-0ab7a3616c32a441a115","guid":"p-isolation-segment-0ab7a3616c32a441a115","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`)),
			}, nil)

			finderOutput, err := service.Find("some-product-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(finderOutput.Product.GUID).To(Equal("some-product-id"))
		})

		Context("failure cases", func() {
			Context("Failed to list staged products", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`%%`)),
					}, nil)

					_, err := service.Find("some-product-name")
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal staged products response")))
				})
			})

			Context("Target product not in staged product list", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`[
					{"installation_name":"cf-15b22d1810a034ea3aca","guid":"cf-15b22d1810a034ea3aca","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"p-isolation-segment-0ab7a3616c32a441a115","guid":"p-isolation-segment-0ab7a3616c32a441a115","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`)),
					}, nil)

					_, err := service.Find("some-product-name")
					Expect(err).To(MatchError(ContainSubstring("could not find product \"some-product-name\"")))
				})
			})
		})
	})
})
