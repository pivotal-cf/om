package api_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/om/api"
)

var _ = Describe("VMTypes", func() {
	var (
		client  *ghttp.Server
		service api.Api
	)

	BeforeEach(func() {
		client = ghttp.NewServer()

		service = api.New(api.ApiInput{
			Client: httpClient{
				client.URL(),
			},
		})
	})

	AfterEach(func() {
		client.Close()
	})

	Context("creating a vm type", func() {
		It("creates a VM Type", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/vm_types"),
					ghttp.VerifyContentType("application/json"),
					ghttp.VerifyJSON(`{ 
					"vm_types": [{
						"name": "my-vm-type",
						"ram": 3840,
						"cpu": 2,
						"ephemeral_disk": 32768    
					}, {
						"name": "my-vm-type2",
						"ram": 3842,
						"cpu": 2,
						"ephemeral_disk": 32790,
						"raw_instance_storage": true
					}] 
				}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

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

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when the http status is non-200 for creating vm types", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v0/vm_types"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			err := service.CreateCustomVMTypes(api.CreateVMTypes{
				VMTypes: []api.CreateVMType{
					api.CreateVMType{Name: "foo"},
				},
			})

			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.Close()

			err := service.CreateCustomVMTypes(api.CreateVMTypes{
				VMTypes: []api.CreateVMType{
					api.CreateVMType{Name: "foo"},
				},
			})

			Expect(err).To(MatchError(ContainSubstring("could not send api request to PUT /api/v0/vm_types")))
		})
	})

	Context("list vm types", func() {
		It("lists VM Types", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/vm_types"),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWith(http.StatusOK, `{
					"vm_types": [{
						"name": "t2.micro",
						"ram": 1024,
						"cpu": 1,
						"ephemeral_disk": 8192,
						"raw_instance_storage": false,
						"builtin": true
					}, {
						"name": "t2.small",
						"ram": 2048,
						"cpu": 1,
						"ephemeral_disk": 8192,
						"raw_instance_storage": false,
						"builtin": true
					}, {
						"name": "t2.medium",
						"ram": 3840,
						"cpu": 1,
						"ephemeral_disk": 32768,
						"raw_instance_storage": true,
						"builtin": true
					}, {
						"name": "c4.large",
						"ram": 3840,
						"cpu": 2,
						"ephemeral_disk": 32768,
						"raw_instance_storage": false,
						"builtin": true
					}]
				}`),
				),
			)

			vmtypes, err := service.ListVMTypes()
			Expect(err).ToNot(HaveOccurred())

			Expect(len(vmtypes)).Should(Equal(4))
			Expect(vmtypes[0].Name).Should(Equal("t2.micro"))
			Expect(vmtypes[1].Name).Should(Equal("t2.small"))
		})

		It("returns an error when the http status is non-200 for listing vm types", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v0/vm_types"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			_, err := service.ListVMTypes()

			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})

		It("returns an error when the api endpoint fails", func() {
			client.Close()

			_, err := service.ListVMTypes()
			Expect(err).To(MatchError(ContainSubstring("could not send api request to GET /api/v0/vm_types")))
		})
	})

	Context("deleting vm Types", func() {
		It("deletes VM Types", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/vm_types"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := service.DeleteCustomVMTypes()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error when the http status is non-200 for deleting vm types", func() {
			client.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v0/vm_types"),
					ghttp.RespondWith(http.StatusTeapot, `{}`),
				),
			)

			err := service.DeleteCustomVMTypes()
			Expect(err).To(MatchError(ContainSubstring("request failed")))
		})
	})

	Context("JSON marshalling", func() {
		It("does not include the builtin key in the ExtraPropeties map of a CreateVMType", func() {
			typeJson := `{"name": "type1", "ram": 2048, "cpu": 2, "ephemeral_disk": 10240, "raw_instance_storage": true, "builtin": true}`

			var vmType api.VMType
			err := json.Unmarshal([]byte(typeJson), &vmType)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmType.BuiltIn).To(BeTrue())
			Expect(vmType.CreateVMType.ExtraProperties).ToNot(HaveKey("builtin"))
		})
	})
})
