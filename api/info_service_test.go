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

var _ = Describe("Info", func() {
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

	It("lists the info", func() {
		client.DoReturns(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{ "info": { "version": "v2.1-build.79" } }`)),
		}, nil)

		info, err := service.Info()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Version).To(Equal("v2.1-build.79"))
	})

	Context("Error Cases", func() {
		It("errors if the API call fails", func() {
			client.DoReturns(nil, errors.New("request failed"))
			info, err := service.Info()
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeZero())
		})

		It("errors if the response is not valid", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(strings.NewReader("")),
			}, nil)
			info, err := service.Info()
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeZero())
		})
	})
})
