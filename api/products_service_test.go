package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProductsService", func() {
	Describe("Upload", func() {
		var (
			client     *fakes.HttpClient
			bar        *fakes.Progress
			liveWriter *fakes.LiveWriter
		)

		BeforeEach(func() {
			client = &fakes.HttpClient{}
			liveWriter = &fakes.LiveWriter{}
			bar = &fakes.Progress{}
		})

		It("makes a request to upload the product to the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				time.Sleep(1 * time.Second)
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader("{}")),
				}, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some other content"))
			service := api.NewProductsService(client, bar, liveWriter)

			output, err := service.Upload(api.UploadProductInput{
				ContentLength: 10,
				Product:       strings.NewReader("some content"),
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.UploadProductOutput{}))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/available_products"))
			Expect(request.ContentLength).To(Equal(int64(10)))
			Expect(request.Header.Get("Content-Type")).To(Equal("some content-type"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal("some other content"))

			newReaderContent, err := ioutil.ReadAll(bar.NewBarReaderArgsForCall(0))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(newReaderContent)).To(Equal("some content"))
			Expect(bar.SetTotalArgsForCall(0)).To(BeNumerically("==", 10))
			Expect(bar.KickoffCallCount()).To(Equal(1))
			Expect(bar.EndCallCount()).To(Equal(1))
		})

		It("logs while waiting for a response from the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/available_products" {
					time.Sleep(5 * time.Second)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("some-installation")),
					}, nil
				}
				return nil, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some-fake-installation"))
			service := api.NewProductsService(client, bar, liveWriter)

			_, err := service.Upload(api.UploadProductInput{
				ContentLength: 10,
				Product:       strings.NewReader("some content"),
				ContentType:   "some content-type",
			})
			Expect(err).NotTo(HaveOccurred())

			By("starting the live log writer")
			Expect(liveWriter.StartCallCount()).To(Equal(1))

			By("writing to the live log writer")
			Expect(liveWriter.WriteCallCount()).To(Equal(5))
			for i := 0; i < 5; i++ {
				Expect(string(liveWriter.WriteArgsForCall(i))).To(ContainSubstring(fmt.Sprintf("%ds elapsed", i+1)))
			}

			By("flushing the live log writer")
			Expect(liveWriter.StopCallCount()).To(Equal(1))
		})

		Context("when an error occurs", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewProductsService(client, bar, liveWriter)

					_, err := service.Upload(api.UploadProductInput{})
					Expect(err).To(MatchError("could not make api request to available_products endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewProductsService(client, bar, liveWriter)

					_, err := service.Upload(api.UploadProductInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

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
				case "/api/v0/available_products":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`[{
							"name":"some-product",
							"product_version":"some-version"
						},
						{
							"name":"some-other-product",
							"product_version":"some-other-version"
						}]`)),
					}
				case "/api/v0/deployed/products":
					resp = &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
					}
				}
				return resp, nil
			}
		})

		It("makes a request to stage the product to the Ops Manager", func() {
			service := api.NewProductsService(client, nil, nil)

			err := service.Stage(api.StageProductInput{
				ProductName:    "some-product",
				ProductVersion: "some-version",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(3))
			By("checking for available products")
			avReq := client.DoArgsForCall(0)
			Expect(avReq.URL.Path).To(Equal("/api/v0/available_products"))

			By("checking for deployed products")
			depReq := client.DoArgsForCall(1)
			Expect(depReq.URL.Path).To(Equal("/api/v0/deployed/products"))

			By("posting to the staged products endpoint with the product name and version")
			stReq := client.DoArgsForCall(2)
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
					case "/api/v0/available_products":
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[{
								"name":"some-product",
								"product_version":"1.1.0"
							},
							{
								"name":"some-product",
								"product_version":"1.0.0"
							}]`)),
						}
					case "/api/v0/deployed/products":
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
								{
									"type":"some-product",
									"guid": "some-deployed-guid",
									"installation_name":"some-deployed-guid"
								},
								{
									"type":"some-other-product",
									"guid": "some-other-deployed-guid",
									"installation_name":"some-other-deployed-guid"
								}]`)),
						}
					}

					return resp, nil
				}
			})

			It("makes a request to stage the product to the Ops Manager", func() {
				service := api.NewProductsService(client, nil, nil)

				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.1.0",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(3))
				By("checking for available products")
				avReq := client.DoArgsForCall(0)
				Expect(avReq.URL.Path).To(Equal("/api/v0/available_products"))

				By("checking for deployed products")
				depReq := client.DoArgsForCall(1)
				Expect(depReq.URL.Path).To(Equal("/api/v0/deployed/products"))

				By("posting to the staged products endpoint with the product name and version")
				stReq := client.DoArgsForCall(2)
				Expect(stReq.URL.Path).To(Equal("/api/v0/staged/products/some-deployed-guid"))
				Expect(stReq.Method).To(Equal("PUT"))
				stReqBody, err := ioutil.ReadAll(stReq.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(stReqBody).To(MatchJSON(`{
							"to_version": "1.1.0"
			}`))
			})

		})

		Context("when the requested product is not available", func() {
			It("returns an error", func() {
				service := api.NewProductsService(client, nil, nil)

				err := service.Stage(api.StageProductInput{
					ProductName:    "some-unavailable-product",
					ProductVersion: "1.2",
				})
				Expect(err).To(MatchError(ContainSubstring("cannot find product some-unavailable-product 1.2")))
			})
		})

		Context("when the requested product version is not available", func() {
			It("returns an error", func() {
				service := api.NewProductsService(client, nil, nil)

				err := service.Stage(api.StageProductInput{
					ProductName:    "some-product",
					ProductVersion: "1.2",
				})
				Expect(err).To(MatchError(ContainSubstring("cannot find product some-product 1.2")))
			})
		})

		Context("when an error occurs", func() {
			Context("when the available products endpoint returns an error", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{})
					Expect(err).To(MatchError("could not make api request to available_products endpoint: some client error"))
				})
			})

			Context("when the available products endpoint returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to available_products endpoint: unexpected response")))
				})
			})

			Context("when the available products endpoint returns invalid JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("%%%")),
					}, nil)
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{})
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})

			Context("when the deployed products endpoint returns an error", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						var err error
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
						}
						switch req.URL.Path {
						case "/api/v0/available_products":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body: ioutil.NopCloser(bytes.NewBufferString(`[{
							"name":"some-product",
							"product_version":"some-version"
						},
						{
							"name":"some-other-product",
							"product_version":"some-other-version"
						}]`)),
							}
						case "/api/v0/deployed/products":
							resp = &http.Response{}
							err = errors.New("some client error")
						}
						return resp, err
					}
				})

				It("returns an error", func() {
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{
						ProductName:    "some-product",
						ProductVersion: "some-version",
					})
					Expect(err).To(MatchError("could not make api request to deployed products endpoint: some client error"))
				})
			})

			Context("when the deployed products endpoint returns a non-200 status code", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						var err error
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
						}
						switch req.URL.Path {
						case "/api/v0/available_products":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body: ioutil.NopCloser(bytes.NewBufferString(`[{
							"name":"some-product",
							"product_version":"some-version"
						},
						{
							"name":"some-other-product",
							"product_version":"some-other-version"
						}]`)),
							}
						case "/api/v0/deployed/products":
							resp = &http.Response{
								Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
								StatusCode: http.StatusInternalServerError,
							}
						}
						return resp, err
					}
				})

				It("returns an error", func() {
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{
						ProductName:    "some-product",
						ProductVersion: "some-version",
					})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to deployed products endpoint: unexpected response")))
				})
			})

			Context("when the deployed products endpoint returns invalid JSON", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						var err error
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
						}
						switch req.URL.Path {
						case "/api/v0/available_products":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body: ioutil.NopCloser(bytes.NewBufferString(`[{
							"name":"some-product",
							"product_version":"some-version"
						},
						{
							"name":"some-other-product",
							"product_version":"some-other-version"
						}]`)),
							}
						case "/api/v0/deployed/products":
							resp = &http.Response{
								StatusCode: http.StatusOK,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`%%%`)),
							}
						}
						return resp, err
					}
				})

				It("returns an error", func() {
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{
						ProductName:    "some-product",
						ProductVersion: "some-version",
					})
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})

			Context("when the staged products endpoint returns an error", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[{
								"name": "foo",
								"product_version": "bar"
							}]`)),
						}
						if req.URL.Path == "/api/v0/staged/products" {
							return nil, fmt.Errorf("some error")
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					service := api.NewProductsService(client, nil, nil)

					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					})
					Expect(err).To(MatchError("could not make api request to staged products endpoint: some error"))
				})
			})

			Context("when the staged products endpoint returns a non-200 status code", func() {
				BeforeEach(func() {
					client = &fakes.HttpClient{}
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						var resp *http.Response
						resp = &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[{
								"name": "foo",
								"product_version": "bar"
							}]`)),
						}
						if req.URL.Path == "/api/v0/staged/products" {
							return &http.Response{
								StatusCode: http.StatusInternalServerError,
								Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
							}, nil
						}
						return resp, nil
					}
				})

				It("returns an error", func() {
					service := api.NewProductsService(client, nil, nil)
					err := service.Stage(api.StageProductInput{
						ProductName:    "foo",
						ProductVersion: "bar",
					})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to staged products endpoint: unexpected response. Please make sure the product you are adding is compatible with everything that is currently staged/deployed.")))
				})
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
			service := api.NewProductsService(client, nil, nil)

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
					service := api.NewProductsService(client, nil, nil)

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
					service := api.NewProductsService(client, nil, nil)

					_, err := service.StagedProducts()
					Expect(err).To(MatchError(ContainSubstring("could not make api request to staged products endpoint: unexpected response")))
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
					service := api.NewProductsService(client, nil, nil)

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
			service := api.NewProductsService(client, nil, nil)

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
			service := api.NewProductsService(client, nil, nil)

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
					service := api.NewProductsService(client, nil, nil)

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
					service := api.NewProductsService(client, nil, nil)

					err := service.Configure(api.ProductsConfigurationInput{
						GUID:          "foo",
						Configuration: `{}`,
					})
					Expect(err).To(MatchError(ContainSubstring("could not make api request to staged product properties endpoint: unexpected response")))
				})
			})
		})
	})
})
