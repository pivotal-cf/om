package api_test

import (
	"errors"
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

			Expect(output).To(Equal(api.ErrandsListOutput{
				Errands: []api.Errand{
					{Name: "first-errand", PostDeploy: "true"},
					{Name: "second-errand", PostDeploy: "false"},
					{Name: "third-errand", PreDelete: "true"},
				},
			}))

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
