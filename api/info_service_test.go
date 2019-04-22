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

var _ = Describe("Info Service", func() {
	Context("VersionAtLeast()", func() {
		It("determines whether a version meets a minimum requirement", func() {
			tests := []struct {
				ver    string
				result bool
			}{
				{"1.2-build10", false},
				{"2.2-build3", true},
				{"1.9-build1", false},
				{"1.12-build1", false},
				{"2.0-build1", false},
				{"2.3-build33", true},
				{"2.1.0-build2", false},
				{"2.2.2-build2", true},
				{"2.5.0-build33", true},
				{"2.5.0", true},
			}
			for _, test := range tests {
				Expect(api.Info{Version: test.ver}.VersionAtLeast(2, 2)).To(Equal(test.result))
			}
		})

		Context("error cases", func() {
			It("returns an error when no version defined", func() {
				ok, err := api.Info{Version: ""}.VersionAtLeast(2, 2)
				Expect(ok).To(Equal(false))
				Expect(err.Error()).To(ContainSubstring("invalid version: ''"))
			})

			It("returns an error when major version is not a number", func() {
				ok, err := api.Info{Version: "xxx.1.0"}.VersionAtLeast(2, 2)
				Expect(ok).To(Equal(false))
				Expect(err.Error()).To(ContainSubstring("invalid version: 'xxx.1.0'"))
			})

			It("returns an error when minor version is not a number", func() {
				ok, err := api.Info{Version: "1.xxx.0"}.VersionAtLeast(2, 2)
				Expect(ok).To(Equal(false))
				Expect(err.Error()).To(ContainSubstring("invalid version: '1.xxx.0'"))
			})
		})
	})

	Context("Info()", func() {
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
})
