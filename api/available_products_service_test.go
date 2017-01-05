package api_test

import (
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

var _ = Describe("AvailableProductsService", func() {
	var (
		client     *fakes.HttpClient
		bar        *fakes.Progress
		liveWriter *fakes.LiveWriter
		service    api.AvailableProductsService
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		liveWriter = &fakes.LiveWriter{}
		bar = &fakes.Progress{}

		service = api.NewAvailableProductsService(client, bar, liveWriter)
	})

	Describe("Upload", func() {
		It("makes a request to upload the product to the Ops Manager", func() {
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader("{}")),
				}, nil
			}

			bar.NewBarReaderReturns(strings.NewReader("some other content"))

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

					_, err := service.Upload(api.UploadProductInput{})
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})
		})
	})

	Describe("List", func() {
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
			output, err := service.List()
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
					_, err := service.List()
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			Context("when the server won't fetch available products", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
					}, nil)

					_, err := service.List()
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})

			Context("when the response is not JSON", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader(`asdf`)),
					}, nil)

					_, err := service.List()
					Expect(err).To(MatchError(ContainSubstring("could not unmarshal")))
				})
			})
		})
	})

	Describe("Trash", func() {
		It("deletes unused products", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`)),
			}, nil)

			err := service.Trash()
			Expect(err).NotTo(HaveOccurred())

			req := client.DoArgsForCall(0)
			Expect(req.URL.Path).To(Equal("/api/v0/available_products"))
			Expect(req.Method).To(Equal("DELETE"))
		})

		Context("failure cases", func() {
			Context("when the client errors before the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some client error"))

					err := service.Trash()
					Expect(err).To(MatchError("could not make api request to available_products endpoint: some client error"))
				})
			})

			Context("when the api returns a non-200 status code", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("{}")),
					}, nil)

					err := service.Trash()
					Expect(err).To(MatchError(ContainSubstring("could not make api request to available_products endpoint: unexpected response")))
				})
			})
		})
	})

	Describe("CheckProductAvailability", func() {
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

		Context("when the product is available", func() {
			It("is true", func() {
				available, err := service.CheckProductAvailability("available-product", "available-version")
				Expect(err).NotTo(HaveOccurred())

				Expect(available).To(BeTrue())
			})
		})

		Context("when the product is unavailable", func() {
			It("is false", func() {
				available, err := service.CheckProductAvailability("unavailable-product", "available-version")
				Expect(err).NotTo(HaveOccurred())

				Expect(available).To(BeFalse())
			})
		})

		Describe("errors", func() {
			Context("the client can't connect to the server", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))
					_, err := service.CheckProductAvailability("", "")
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})
		})
	})
})
