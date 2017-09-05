package api_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("ErrandsService", func() {
	var (
		client  *fakes.HttpClient
		service api.ErrandsService
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.NewErrandsService(client)
	})

	Describe("SetState", func() {
		It("sets state for a product's errands", func() {
			var path, method string
			var header http.Header
			var body io.Reader
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				path = req.URL.Path
				method = req.Method
				body = req.Body
				header = req.Header

				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader("{}")),
				}, nil
			}

			err := service.SetState("some-product-id", "some-errand", "when-changed", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal("/api/v0/staged/products/some-product-id/errands"))
			Expect(method).To(Equal("PUT"))
			Expect(header.Get("Content-Type")).To(Equal("application/json"))

			bodyBytes, err := ioutil.ReadAll(body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(MatchJSON(`{
				"errands": [
            {
              "name": "some-errand",
              "post_deploy": "when-changed",
              "pre_delete": false
            }
					]
			}`))
		})

		Context("failure cases", func() {
			Context("when ops manager returns a not-OK response code", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusTeapot,
							Body: ioutil.NopCloser(strings.NewReader("I'm a teapot")),
						}, nil
					}

					err := service.SetState("some-product-id", "some-errand", "when-changed", "false")
					Expect(err).To(MatchError("failed to set errand state: 418 I'm a teapot"))
				})
			})

			Context("when the product ID cannot be URL encoded", func() {
				It("returns an error", func() {
					err := service.SetState("%%%", "some-errand", "true", "false")
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			Context("when the client cannot make a request", func() {
				It("returns an error", func() {
					client.DoReturns(nil, errors.New("client do errored"))

					err := service.SetState("some-product-id", "some-errand", "true", "false")
					Expect(err).To(MatchError("client do errored"))
				})
			})

			Context("when the response body cannot be read", func() {
				BeforeEach(func() {
					api.SetReadAll(func(_ io.Reader) ([]byte, error) {
						return nil, errors.New("failed to read body")
					})
				})

				AfterEach(func() {
					api.ResetReadAll()
				})

				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusTeapot,
							Body: nil,
						}, nil
					}

					err := service.SetState("some-product-id", "some-errand", "true", "false")
					Expect(err).To(MatchError(ContainSubstring("failed to read body")))
				})
			})
		})
	})

	Describe("List", func() {
		It("lists errands for a product", func() {
			var path string
			client.DoStub = func(req *http.Request) (*http.Response, error) {
				path = req.URL.Path

				return &http.Response{StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"errands": [
								{"post_deploy":"true","name":"first-errand"},
								{"post_deploy":"false","name":"second-errand"},
								{"pre_delete":"true","name":"third-errand"}
							]
						}`)),
				}, nil
			}

			output, err := service.List("some-product-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(output.Errands).To(ConsistOf([]api.Errand{
				{Name: "first-errand", PostDeploy: "true"},
				{Name: "second-errand", PostDeploy: "false"},
				{Name: "third-errand", PreDelete: "true"},
			},
			))

			Expect(path).To(Equal("/api/v0/staged/products/some-product-id/errands"))
		})

		Context("failure cases", func() {
			Context("when the product ID cannot be URL encoded", func() {
				It("returns an error", func() {
					_, err := service.List("%%%")
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})

			Context("when the client cannot make a request", func() {
				It("returns an error", func() {
					client.DoReturns(nil, errors.New("client do errored"))

					_, err := service.List("some-product-id")
					Expect(err).To(MatchError("client do errored"))
				})
			})

			Context("when the response body cannot be parsed", func() {
				It("returns an error", func() {
					client.DoStub = func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(strings.NewReader(`%%%%`)),
						}, nil
					}

					_, err := service.List("some-product-id")
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
