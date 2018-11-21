package api_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

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

	Describe("CheckProductAvailability", func() {
		BeforeEach(func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`[{
						"name": "available-product",
						"product_version": "available-version"
					}]`)),
			}, nil)
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

		Context("When an error occurs", func() {
			Context("when the client can't connect to the server", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{}, errors.New("some error"))
					_, err := service.CheckProductAvailability("", "")
					Expect(err).To(MatchError(ContainSubstring("could not make api request")))
				})
			})
		})
	})

	Describe("RunningInstallation", func() {
		It("returns only the running installation on the Ops Manager", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"finished_at": null,
							"status": "running",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"status": "succeeded",
							"id": 2
						}
					]
				}`))}, nil)

			output, err := service.RunningInstallation()

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(api.InstallationsServiceOutput{
				ID:         3,
				Status:     "running",
				FinishedAt: parseTime(nil),
			}))

			req := client.DoArgsForCall(0)

			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/api/v0/installations"))
		})

		Context("when there are no installations", func() {
			It("returns a zero value installation", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": []}`))}, nil)

				output, err := service.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))
			})
		})

		Context("when there is no running installation", func() {
			It("returns a zero value installation", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"user_name": "admin",
							"finished_at": "2017-05-25T00:10:00.303Z",
							"status": "succeeded",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-05-24T23:55:56.106Z",
							"status": "succeeded",
							"id": 2
						}
					]
				}`))}, nil)

				output, err := service.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))

			})
		})

		Context("when only an earlier installation is listed in the running state", func() {
			It("does not consider the earlier installation to be running", func() {
				client.DoReturns(&http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
					"installations": [
						{
							"finished_at": null,
							"status": "succeeded",
							"id": 3
						},
						{
							"user_name": "admin",
							"finished_at": "2017-07-05T00:39:32.123Z",
							"status": "running",
							"id": 2
						}
					]
				}`))}, nil)

				output, err := service.RunningInstallation()

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(api.InstallationsServiceOutput{}))

				req := client.DoArgsForCall(0)

				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.Path).To(Equal("/api/v0/installations"))
			})
		})

		Context("error cases", func() {
			Context("when the client has an error during the request", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, errors.New("some error"))

					_, err := service.RunningInstallation()
					Expect(err).To(MatchError("could not make api request to installations endpoint: could not send api request to GET /api/v0/installations: some error"))
				})
			})

			Context("when the client returns a non-2XX", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					}, nil)

					_, err := service.RunningInstallation()
					Expect(err).To(MatchError(ContainSubstring("request failed: unexpected response")))
				})
			})

			Context("when the json cannot be decoded", func() {
				It("returns an error", func() {
					client.DoReturns(&http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(strings.NewReader("##################")),
					}, nil)

					_, err := service.RunningInstallation()
					Expect(err).To(MatchError(ContainSubstring("failed to decode response: invalid character")))
				})
			})
		})
	})

	Describe("GetStagedProductByName", func() {
		It("Find product by product name", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewBufferString(`[
					{"installation_name":"p-bosh","guid":"some-product-id","type":"some-product-name","product_version":"1.10.0.0"},
					{"installation_name":"cf-15b22d1810a034ea3aca","guid":"cf-15b22d1810a034ea3aca","type":"cf","product_version":"1.10.0-build.177"},
					{"installation_name":"p-isolation-segment-0ab7a3616c32a441a115","guid":"p-isolation-segment-0ab7a3616c32a441a115","type":"p-isolation-segment","product_version":"1.10.0-build.31"}
				]`)),
			}, nil)

			finderOutput, err := service.GetStagedProductByName("some-product-name")
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

					_, err := service.GetStagedProductByName("some-product-name")
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

					_, err := service.GetStagedProductByName("some-product-name")
					Expect(err).To(MatchError(ContainSubstring("could not find product \"some-product-name\"")))
				})
			})
		})
	})
})
