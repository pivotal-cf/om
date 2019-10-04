package api_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Available Products", func() {
	var (
		progressClient *fakes.HttpClient
		client         *fakes.HttpClient
		service        api.Api
	)

	BeforeEach(func() {
		progressClient = &fakes.HttpClient{}
		client = &fakes.HttpClient{}

		service = api.New(api.ApiInput{
			Client:         client,
			ProgressClient: progressClient,
		})
	})

	Describe("UploadAvailableProduct", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			progressClient.DoStub = func(req *http.Request) (*http.Response, error) {
				Expect(req.Context().Value("polling-interval")).To(Equal(time.Second))

				body, err := ioutil.ReadAll(req.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(Equal("some content"))

				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil
			}

			output, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
				ContentLength:   10,
				Product:         strings.NewReader("some content"),
				ContentType:     "some content-type",
				PollingInterval: 1,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.UploadAvailableProductOutput{}))

			request := progressClient.DoArgsForCall(0)
			Expect(request.Method).To(Equal("POST"))
			Expect(request.URL.Path).To(Equal("/api/v0/available_products"))
			Expect(request.ContentLength).To(Equal(int64(10)))
			Expect(request.Header.Get("Content-Type")).To(Equal("some content-type"))
		})

		When("an error occurs", func() {
			When("the client errors performing the request", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{}, errors.New("some client error"))

					_, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError("could not make api request to available_products endpoint: some client error"))
				})
			})

			When("the api returns a non-200 status code", func() {
				It("returns an error", func() {
					progressClient.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.UploadAvailableProduct(api.UploadAvailableProductInput{
						PollingInterval: 1,
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("ListAvailableProducts", func() {
		BeforeEach(func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/api/v0/available_products" {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(strings.NewReader(`[{
						"name": "available-product",
						"product_version": "available-version"
					}]`)),
					}, nil
				}
				return nil, nil
			}
		})

		It("lists available products", func() {
			output, err := service.ListAvailableProducts()
			Expect(err).NotTo(HaveOccurred())

			Expect(output.ProductsList).To(ConsistOf(
				[]api.ProductInfo{
					api.ProductInfo{
						Name:    "available-product",
						Version: "available-version",
					},
				}))
		})

		Describe("errors", func() {
			Context("the client can't connect to the server", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))
					_, err := service.ListAvailableProducts()
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			When("the server won't fetch available products", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.ListAvailableProducts()
					Expect(err).To(MatchError(ContainSubstring("request failed")))
				})
			})

			When("the response is not JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`asdf`)),
					}, nil)

					_, err := service.ListAvailableProducts()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
				})
			})
		})
	})

	Describe("DeleteAvailableProducts", func() {
		It("deletes a named product / version", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			}, nil)

			err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
				ProductName:    "some-product",
				ProductVersion: "1.2.3-build.4",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DoCallCount()).To(Equal(1))

			request := client.DoArgsForCall(0)
			Expect(request.Method).To(Equal("DELETE"))
			Expect(request.URL.Path).To(Equal("/api/v0/available_products"))
			Expect(request.URL.RawQuery).To(Equal("product_name=some-product&version=1.2.3-build.4"))
		})

		When("the ShouldDeleteAllProducts flag is provided", func() {
			It("does not provide a product query to DELETE", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
				}, nil)

				err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
					ShouldDeleteAllProducts: true,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DoCallCount()).To(Equal(1))

				req := client.DoArgsForCall(0)
				Expect(req.Method).To(Equal("DELETE"))
				Expect(req.URL.Path).To(Equal("/api/v0/available_products"))
				Expect(req.URL.RawQuery).To(Equal(""))
			})
		})

		When("an error occurs", func() {
			When("a non-200 status code is returned", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					err := service.DeleteAvailableProducts(api.DeleteAvailableProductsInput{
						ProductName:    "some-product",
						ProductVersion: "1.2.3-build.4",
					})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response:")))
				})
			})
		})
	})
})
