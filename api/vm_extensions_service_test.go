package api_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
)

var _ = Describe("VMExtensions", func() {
	var (
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			Client: client,
		})

		client.DoReturns(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)
	})

	It("creates a VM Extension", func() {
		err := service.CreateStagedVMExtension(api.CreateVMExtension{
			Name:            "some-vm-extension",
			CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
		})

		Expect(err).NotTo(HaveOccurred())

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)

		Expect(req.Method).To(Equal("PUT"))
		Expect(req.URL.Path).To(Equal("/api/v0/staged/vm_extensions/some-vm-extension"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

		jsonBody, err := ioutil.ReadAll(req.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(jsonBody).To(MatchJSON(`{
			"name": "some-vm-extension",
			"cloud_properties": {"iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"]}
		}`))
	})

	Context("failure cases", func() {
		It("returns an error when the http status is non-200", func() {

			client.DoReturns(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := service.CreateStagedVMExtension(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			})

			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

			err := service.CreateStagedVMExtension(api.CreateVMExtension{
				Name:            "some-vm-extension",
				CloudProperties: json.RawMessage(`{ "iam_instance_profile": "some-iam-profile", "elbs": ["some-elb"] }`),
			})

			Expect(err).To(MatchError("could not send api request to PUT /api/v0/staged/vm_extensions/some-vm-extension: api endpoint failed"))
		})
	})
})
