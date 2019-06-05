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

var _ = Describe("VMTypes", func() {
	var (
		client  *fakes.HttpClient
		service api.Api
	)

	BeforeEach(func() {
		client = &fakes.HttpClient{}
		service = api.New(api.ApiInput{
			Client: client,
		})
		client.DoStub = func(req *http.Request) (*http.Response, error) {
			if req.Method == "GET" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(strings.NewReader(
						`{
					"vm_types": [
						{
						"name": "t2.micro",
						"ram": 1024,
						"cpu": 1,
						"ephemeral_disk": 8192,
						"raw_instance_storage": false,
						"builtin": true
						},
						{
						"name": "t2.small",
						"ram": 2048,
						"cpu": 1,
						"ephemeral_disk": 8192,
						"raw_instance_storage": false,
						"builtin": true
						},
						{
						"name": "t2.medium",
						"ram": 3840,
						"cpu": 1,
						"ephemeral_disk": 32768,
						"raw_instance_storage": true,
						"builtin": true
						},
						{
						"name": "c4.large",
						"ram": 3840,
						"cpu": 2,
						"ephemeral_disk": 32768,
						"raw_instance_storage": false,
						"builtin": true
						}
					]
					}
					`))}, nil
			} else {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil
			}
		}
	})

	It("creates a VM Type", func() {
		err := service.CreateCustomVMTypes(api.CreateVMTypes{
			VMTypes: []api.CreateVMType{
				api.CreateVMType{
					Name:          "my-vm-type",
					RAM:           3840,
					CPU:           2,
					EphemeralDisk: 32768,
				},
				api.CreateVMType{Name: "my-vm-type2",
					RAM:             3842,
					CPU:             2,
					EphemeralDisk:   32790,
					ExtraProperties: map[string]interface{}{"raw_instance_storage": true},
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)

		Expect(req.Method).To(Equal("PUT"))
		Expect(req.URL.Path).To(Equal("/api/v0/vm_types"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

		jsonBody, err := ioutil.ReadAll(req.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(jsonBody).To(MatchJSON(`{ 
		"vm_types": [ 
			{
			"name": "my-vm-type",
			"ram": 3840,
			"cpu": 2,
			"ephemeral_disk": 32768    
			},
			{
			"name": "my-vm-type2",
			"ram": 3842,
			"cpu": 2,
			"ephemeral_disk": 32790,
			"raw_instance_storage": true
			}
	  	] 
	  }`))
	})

	It("lists VM Types", func() {
		vmtypes, err := service.ListVMTypes()

		Expect(err).NotTo(HaveOccurred())

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)

		Expect(req.Method).To(Equal("GET"))
		Expect(req.URL.Path).To(Equal("/api/v0/vm_types"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

		Expect(len(vmtypes)).Should(Equal(4))
		Expect(vmtypes[0].Name).Should(Equal("t2.micro"))
		Expect(vmtypes[1].Name).Should(Equal("t2.small"))
	})

	It("deletes VM Types", func() {
		err := service.DeleteCustomVMTypes()
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)

		Expect(req.Method).To(Equal("DELETE"))
		Expect(req.URL.Path).To(Equal("/api/v0/vm_types"))
		Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
	})

	Context("failure cases", func() {
		It("returns an error when the http status is non-200", func() {

			client.DoReturns(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := service.CreateCustomVMTypes(api.CreateVMTypes{
				VMTypes: []api.CreateVMType{
					api.CreateVMType{Name: "foo"},
				},
			})

			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
		It("returns an error when the http status is non-200 for listing vm extensions", func() {

			client.DoReturns(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			_, err := service.ListVMTypes()

			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})
		It("returns an error when the http status is non-200 for deleting vm extensions", func() {

			client.DoReturns(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, nil)

			err := service.DeleteCustomVMTypes()
			Expect(err).To(MatchError(ContainSubstring("500 Internal Server Error")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.DoReturns(&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(`{}`))}, errors.New("api endpoint failed"))

			err := service.CreateCustomVMTypes(api.CreateVMTypes{
				VMTypes: []api.CreateVMType{
					api.CreateVMType{Name: "foo"},
				},
			})

			Expect(err).To(MatchError("could not send api request to PUT /api/v0/vm_types: api endpoint failed"))
		})
	})

	Context("JSON marshalling", func() {
		It("does not include the builtin key in the ExtraPropeties map of a CreateVMType", func() {
			typeJson := `{"name": "type1", "ram": 2048, "cpu": 2, "ephemeral_disk": 10240, "raw_instance_storage": true, "builtin": true}`

			var vmType api.VMType
			err := json.Unmarshal([]byte(typeJson), &vmType)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmType.BuiltIn).To(BeTrue())
			Expect(vmType.CreateVMType.ExtraProperties).NotTo(HaveKey("builtin"))
		})
	})
})
